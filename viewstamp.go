package main

import (
	"io/ioutil"
	"runtime"
	"sync"

	"./labrpc"
	"github.com/tidwall/sjson"

	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	NORMAL = iota
	VIEWCHANGE
	RECOVERING
)

const MAX_RETRIES = 50

type PBWriteAuthServer struct {
	mu             sync.Mutex          // Lock to protect shared access to this peer's state
	peers          []*labrpc.ClientEnd // RPC end points of all peers
	me             int                 // this peer's index into peers[]
	currentView    int                 // what this peer believes to be the current active view
	status         int                 // the server's current status (NORMAL, VIEWCHANGE or RECOVERING)
	lastNormalView int                 // the latest view which had a NORMAL status

	log          []string // the log of "commands"
	commitIndex  int      // all log entries <= commitIndex are considered to have been committed.
	numOfRetries int
}

type PrepareWriteAuthArgs struct {
	View          int    // the primary's current view
	PrimaryCommit int    // the primary's commitIndex
	Index         int    // the index position at which the log entry is to be replicated on backups
	Entry         string // the log entry to be replicated
}

// PrepareReply defines the reply for the Prepare RPC
// Note that all field names must start with a capital letter for an RPC reply struct
type PrepareReply struct {
	View    int  // the backup's current view
	Success bool // whether the Prepare request has been accepted or rejected
}

// RecoverArgs defined the arguments for the Recovery RPC
type RecoveryArgs struct {
	View   int // the view that the backup would like to synchronize with
	Server int // the server sending the Recovery RPC (for debugging)
}

type RecoveryReply struct {
	View          int           // the view of the primary
	Entries       []interface{} // the primary's log including entries replicated up to and including the view.
	PrimaryCommit int           // the primary's commitIndex
	Success       bool          // whether the Recovery request has been accepted or rejected
}

type ViewChangeArgs struct {
	View int // the new view to be changed into
}

type ViewWAChangeReply struct {
	LastNormalView int      // the latest view which had a NORMAL status at the server
	Log            []string // the log at the server
	Success        bool     // whether the ViewChange request has been accepted/rejected
}

type StartWAViewArgs struct {
	View int      // the new view which has completed view-change
	Log  []string // the log associated with the new new
}

type StartWAViewReply struct {
}

func randstring(n int) string {
	b := make([]byte, 2*n)
	crand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	return s[0:n]
}

type configWA struct {
	mu        sync.Mutex
	net       *labrpc.Network
	n         int
	done      int32 // tell internal threads to die
	pbservers []*PBWriteAuthServer
	applyErr  []string   // from apply channel readers
	connected []bool     // whether each server is on the net
	endnames  [][]string // the port file names each sends to
}

var ncpu_once sync.Once

// n is the total number of servers

func make_config_write_auth(n int, unreliable bool) *configWA {
	ncpu_once.Do(func() {
		if runtime.NumCPU() < 2 {
			fmt.Printf("warning: only one CPU, which may conceal locking bugs\n")
		}
	})
	runtime.GOMAXPROCS(4)
	cfg := &configWA{}
	cfg.net = labrpc.MakeNetwork()
	cfg.n = n
	cfg.applyErr = make([]string, cfg.n)
	cfg.pbservers = make([]*PBWriteAuthServer, cfg.n)
	cfg.connected = make([]bool, cfg.n)
	cfg.endnames = make([][]string, cfg.n)

	cfg.setunreliable(unreliable)

	cfg.net.LongDelays(true)

	// create a full set of PBServers
	for i := 0; i < cfg.n; i++ {
		cfg.start1(i)
	}

	// connect everyone
	for i := 0; i < cfg.n; i++ {
		cfg.connect(i)
	}

	return cfg
}

// shut down a server but save its persistent state.
func (cfg *configWA) crash1(i int) {
	cfg.disconnect(i)
	cfg.net.DeleteServer(i) // disable client connections to the server.

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	srv := cfg.pbservers[i]
	if srv != nil {
		cfg.mu.Unlock()
		// srv.Kill()
		cfg.mu.Lock()
		cfg.pbservers[i] = nil
	}
}

// make and initialize a server.
func (cfg *configWA) cleanup() {
	for i := 0; i < len(cfg.pbservers); i++ {
		if cfg.pbservers[i] != nil {
			// cfg.pbservers[i].Kill()
		}
	}
	atomic.StoreInt32(&cfg.done, 1)
}

func (cfg *configWA) start1(i int) {
	cfg.crash1(i)

	// a fresh set of outgoing ClientEnd names.
	// so that old crashed instance's ClientEnds can't send.
	cfg.endnames[i] = make([]string, cfg.n)
	for j := 0; j < cfg.n; j++ {
		cfg.endnames[i][j] = randstring(20)
	}

	// a fresh set of ClientEnds.
	ends := make([]*labrpc.ClientEnd, cfg.n)
	for j := 0; j < cfg.n; j++ {
		ends[j] = cfg.net.MakeEnd(cfg.endnames[i][j])
		cfg.net.Connect(cfg.endnames[i][j], j)
	}

	peer := MakeWA(ends, i, 0)

	cfg.mu.Lock()
	cfg.pbservers[i] = peer
	cfg.mu.Unlock()

	svc := labrpc.MakeService(peer)
	srv := labrpc.MakeServer()
	srv.AddService(svc)
	cfg.net.AddServer(i, srv)
}

// attach server i to the net.
func (cfg *configWA) connect(i int) {
	// fmt.Printf("connect(%d)\n", i)

	cfg.connected[i] = true

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, true)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, true)
		}
	}
}

// detach server i from the net.
func (cfg *configWA) disconnect(i int) {
	// fmt.Printf("disconnect(%d)\n", i)

	cfg.connected[i] = false

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[i] != nil {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, false)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[j] != nil {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, false)
		}
	}
}

func (cfg *configWA) rpcCount(server int) int {
	return cfg.net.GetCount(server)
}
func (cfg *configWA) setunreliable(unrel bool) {
	cfg.net.Reliable(!unrel)
}
func (cfg *configWA) setlongreordering(longrel bool) {
	cfg.net.LongReordering(longrel)
}
func (cfg *configWA) waitCommitted(primary, index int) {
	pri := cfg.pbservers[primary]

	t0 := time.Now()
	for time.Since(t0).Seconds() < 10 {
		committed := pri.IsCommitted(index)
		if committed {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	// // cfg.t.Fatalf("timed out waiting for index %d to commit\n", index)
}
func (cfg *configWA) checkCommittedIndex(index int, command interface{}, expectedServers int) {
	cmd, cmdInitialized := command.(int)

	// check that sufficient number of backups committed the same command at index
	nReplicated := 0
	for i := 0; i < len(cfg.pbservers); i++ {
		if cfg.connected[i] {
			ok, command1 := cfg.pbservers[i].GetEntryAtIndex(index)
			if ok {
				cmd1, ok1 := command1.(int)
				if !ok1 {
					continue
				}
				if !cmdInitialized {
					cmd = cmd1
				}
				if cmd1 == cmd {
					nReplicated++
				}
			}
		}
	}
	if nReplicated < expectedServers {
		// cfg.t.Fatalf("command %v replicated to %d servers, expected replication to %d servers\n", command, nReplicated, expectedServers)
	}
}
func (cfg *configWA) replicateWriteAuth(server int, username string, authToken string, expectedServers int) (
	index int) {
	// submit command to primary
	var pri *PBWriteAuthServer
	// cfg.mu.Lock()
	pri = cfg.pbservers[server]
	// cfg.mu.Unlock()
	index, _, ok := pri.StartWriteAuth(username, authToken)
	if !ok {
		// cfg.t.Fatalf("node-%d rejected command\n", server)
	}
	// primary submitted our request, wait for a while for
	// it to be considered committed and check that it has
	// replicated to sufficient number of servers
	t0 := time.Now()
	for time.Since(t0).Seconds() < 10 {

		committed := pri.IsCommitted(index)
		if committed {

			nReplicated := 0
			for i := 0; i < len(cfg.pbservers); i++ {
				ok, cmd1 := cfg.pbservers[i].GetWriteAuthEntryAtIndex(index)
				if ok {
					if cmd1 == username+","+authToken {
						nReplicated++
					}
				}
			}
			if nReplicated >= expectedServers {
				return index
			}
		}
		time.Sleep(80 * time.Millisecond)
	}
	// cfg.t.Fatalf("timed out replicating cmd (%v) to %d expected servers\n", cmd, expectedServers)
	return -1
}
func (cfg *configWA) viewChange(newView int) {
	primary := GetPrimary(newView, len(cfg.pbservers))
	if !cfg.connected[primary] {
		// cfg.t.Fatalf("node-%d not connected to perform view change\n", primary)
	}
	pri := cfg.pbservers[primary]
	pri.PromptViewChange(newView)
	t0 := time.Now()
	for time.Since(t0).Seconds() < 10 {
		view, ok := pri.ViewStatus()
		if ok && view >= newView {
			return
		}
		time.Sleep(80 * time.Millisecond)
	}
	// cfg.t.Fatalf("timed out waiting for view %d change to complete at node-%d\n", newView, primary)
}

// func writeAuthsReplicated(username string, authToken string) {
// 	for index := 1; index <= 10; index++ {
// 		xindex := cfg.replicateOne(primaryID, 1000+index, servers) // replicate command 1000+index, expected successful replication to all servers
// 	}
// }

// GetPrimary is an auxilary function that returns the server index of the
// primary server given the view number (and the total number of replica servers)
func GetPrimary(view int, nservers int) int {
	fmt.Println(view, nservers)
	return view % nservers
}

// IsCommitted is called by tester to check whether an index position
// has been considered committed by this server
func (srv *PBWriteAuthServer) IsCommitted(index int) bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.commitIndex >= index {
		return true
	}
	return false
}

// ViewStatus is called by tester to find out the current view of this server
// and whether this view has a status of NORMAL.
func (srv *PBWriteAuthServer) ViewStatus() (currentView int, statusIsNormal bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.currentView, srv.status == NORMAL
}

// GetEntryAtIndex is called by tester to return the command replicated at
// a specific log index. If the server's log is shorter than "index", then
// ok = false, otherwise, ok = true
func (srv *PBWriteAuthServer) GetEntryAtIndex(index int) (ok bool, command interface{}) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.log) > index {
		return true, srv.log[index]
	}
	return false, command
}
func (srv *PBWriteAuthServer) GetWriteAuthEntryAtIndex(index int) (ok bool, user interface{}) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.log) > index {
		return true, srv.log[index]
	}
	return false, user
}

// Kill is called by tester to clean up (e.g. stop the current server)
// before moving on to the next test
// func (srv *PBServer) Kill() {
// 	// Your code here, if necessary
// }

// Make is called by tester to create and initalize a PBServer
// peers is the list of RPC endpoints to every server (including self)
// me is this server's index into peers.
// startingView is the initial view (set to be zero) that all servers start in
func MakeWA(peers []*labrpc.ClientEnd, me int, startingView int) *PBWriteAuthServer {
	srv := &PBWriteAuthServer{
		peers:          peers,
		me:             me,
		currentView:    startingView,
		lastNormalView: startingView,
		status:         NORMAL,
	}
	// all servers' log are initialized with a dummy command at index 0
	var user string
	srv.log = append(srv.log, user)

	// Your other initialization code here, if there's any
	return srv
}

// Start() is invoked by tester on some replica server to replicate a
// command.  Only the primary should process this request by appending
// the command to its log and then return *immediately* (while the log is being replicated to backup servers).
// if this server isn't the primary, returns false.
// Note that since the function returns immediately, there is no guarantee that this command
// will ever be committed upon return, since the primary
// may subsequently fail before replicating the command to all servers
//
// The first return value is the index that the command will appear at
// *if it's eventually committed*. The second return value is the current
// view. The third return value is true if this server believes it is
// the primary.
func (srv *PBWriteAuthServer) StartWriteAuth(username string, authToken string) (
	index int, view int, ok bool) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	// do not process command if status is not NORMAL
	// and if i am not the primary in the current view
	if srv.status != NORMAL {
		return -1, srv.currentView, false
	} else if GetPrimary(srv.currentView, len(srv.peers)) != srv.me {
		return -1, srv.currentView, false
	}

	srv.log = append(srv.log, authToken+","+username)
	index = len(srv.log) - 1
	view = srv.currentView
	repliesChannel := make(chan bool)

	setupWriteAuthPrepareCalls(srv, repliesChannel, view, username, authToken, index, srv.commitIndex)

	ok = true

	return index, view, ok
}

func setupWriteAuthPrepareCalls(srv *PBWriteAuthServer, repliesChannel chan bool, currentView int, username string, authToken string, currentIndex int, commitIndex int) {

	for peer := range srv.peers {

		if peer != srv.me {

			go func(peer int) {
				var response PrepareReply
				repliesChannel <- srv.sendWriteAuthPrepare(peer, &PrepareWriteAuthArgs{View: currentView, PrimaryCommit: commitIndex, Index: currentIndex, Entry: authToken + "," + username}, &response)
			}(peer)

		}
	}

	go checkWriteAuthReplies(srv, repliesChannel, currentView, username, authToken, currentIndex, commitIndex)
}

func checkWriteAuthReplies(srv *PBWriteAuthServer, repliesChannel chan bool, currentView int, username string, authToken string, currentIndex int, commitIndex int) {

	successCount := 0
	totalCount := 0

	for reply := range repliesChannel {

		totalCount++

		if reply == true {
			successCount++
		}

		if successCount == len(srv.peers)/2 {
			break
		} else if totalCount == len(srv.peers)-1 {
			break
		}
	}

	if successCount >= len(srv.peers)/2 {

		committed := false

		for {

			srv.mu.Lock()
			if srv.commitIndex < currentIndex {
				srv.commitIndex = currentIndex
				committed = true
			}

			srv.mu.Unlock()

			if committed {
				fmt.Println("Committing value")
				authsData, err := ioutil.ReadFile(authsFilePath + "_" + string(srv.me))
				modifiedJson, err := sjson.Set(string(authsData), authToken, username)
				checkErr(err)
				ioutil.WriteFile(authsFilePath+"_"+string(srv.me), []byte(modifiedJson), 0644)
				break
			}
		}

	} else { //Failure Condition. Retry Again
		go setupWriteAuthPrepareCalls(srv, repliesChannel, currentView, username, authToken, currentIndex, commitIndex)
	}
}

// exmple code to send an AppendEntries RPC to a server.
// server is the index of the target server in srv.peers[].
// expects RPC arguments in args.
// The RPC library fills in *reply with RPC reply, so caller should pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
func (srv *PBWriteAuthServer) sendWriteAuthPrepare(server int, args *PrepareWriteAuthArgs, reply *PrepareReply) bool {
	ok := srv.peers[server].Call("PBWriteAuthServer.Prepare", args, reply)
	return ok
}

// Some external oracle prompts the primary of the newView to
// switch to the newView.
// PromptViewChange just kicks start the view change protocol to move to the newView
// It does not block waiting for the view change process to complete.
func (srv *PBWriteAuthServer) PromptViewChange(newView int) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	newPrimary := GetPrimary(newView, len(srv.peers))

	if newPrimary != srv.me { //only primary of newView should do view change
		return
	} else if newView <= srv.currentView {
		return
	}
	vcArgs := &ViewChangeArgs{
		View: newView,
	}
	vcReplyChan := make(chan *ViewWAChangeReply, len(srv.peers))
	// send ViewChange to all servers including myself
	for i := 0; i < len(srv.peers); i++ {
		go func(server int) {
			var reply ViewWAChangeReply
			ok := srv.peers[server].Call("PBServer.ViewChange", vcArgs, &reply)
			// fmt.Printf("node-%d (nReplies %d) received reply ok=%v reply=%v\n", srv.me, nReplies, ok, r.reply)
			if ok {
				vcReplyChan <- &reply
			} else {
				vcReplyChan <- nil
			}
		}(i)
	}

	// wait to receive ViewChange replies
	// if view change succeeds, send StartView RPC
	go func() {
		var successReplies []*ViewWAChangeReply
		var nReplies int
		majority := len(srv.peers)/2 + 1
		for r := range vcReplyChan {
			nReplies++
			if r != nil && r.Success {
				successReplies = append(successReplies, r)
			}
			if nReplies == len(srv.peers) || len(successReplies) == majority {
				break
			}
		}
		ok, log := srv.determineNewViewLog(successReplies)
		if !ok {
			return
		}
		svArgs := &StartWAViewArgs{
			View: vcArgs.View,
			Log:  log,
		}
		// send StartView to all servers including myself
		for i := 0; i < len(srv.peers); i++ {
			var reply StartWAViewReply
			go func(server int) {
				// fmt.Printf("node-%d sending StartView v=%d to node-%d\n", srv.me, svArgs.View, server)
				srv.peers[server].Call("PBServer.StartView", svArgs, &reply)
			}(i)
		}
	}()
}

func (srv *PBWriteAuthServer) determineNewViewLog(successReplies []*ViewWAChangeReply) (
	ok bool, newViewLog []string) {

	latestAccessedView := -999
	ok = false

	if len(successReplies) > len(srv.peers)/2 {

		for _, reply := range successReplies {

			if reply.LastNormalView > latestAccessedView {

				latestAccessedView = reply.LastNormalView
				newViewLog = reply.Log

			} else if reply.LastNormalView == latestAccessedView {

				if len(reply.Log) > len(newViewLog) {
					newViewLog = reply.Log
				}
			}
		}
		ok = true
	}

	return ok, newViewLog
}

// ViewChange is the RPC handler to process ViewChange RPC.
func (srv *PBWriteAuthServer) ViewChange(args *ViewChangeArgs, reply *ViewWAChangeReply) {

	reply.Success = false
	srv.mu.Lock()

	if srv.currentView < args.View {

		reply.Log = srv.log
		reply.LastNormalView = srv.lastNormalView
		reply.Success = true
		srv.status = VIEWCHANGE

	}

	srv.mu.Unlock()
}

// StartView is the RPC handler to process StartView RPC.
func (srv *PBWriteAuthServer) StartView(args *StartWAViewArgs, reply *StartWAViewReply) {

	srv.mu.Lock()

	if srv.currentView <= args.View {

		srv.commitIndex = len(srv.log) - 1
		srv.log = args.Log
		srv.lastNormalView = args.View
		srv.currentView = args.View
		srv.status = NORMAL

	}

	srv.mu.Unlock()
}
