import React, {Component} from 'react';
import FlatButton from 'material-ui/FlatButton';
import TextField from 'material-ui/TextField';
import Dialog from 'material-ui/Dialog';

import * as request from 'axios';

const apiBasePath = 'http://localhost:8080/';
const endpoints = {
    "register" : 'register',
    "login" : 'login',
};

class LoginRegister extends Component {
    constructor(props){
        super(props)
        this.state = {
            "dialogOpen": true,
            "username": "",
            "password": "",
            "name": "",
            "email": "",
            "confirm": "",
        }
    }

    componentWillUpdate(){
        if(this.state.dialogOpen === false){
            this.setState({
                dialogOpen: true
            })
        }
    }

    handleSubmit(){
        let body = new FormData();
        const { username, password, name, email, confirm } = this.state;
        if(this.props.dialog === "register"){
            if(name === "" || username === "" || password === "" || confirm === ""){
                return alert("Please Enter All Required Fields!");
            }
            else if(password !== confirm){                
                return alert("Password and Confirm Password Fields must match!");
            }
            body.set('name', name);
            body.set('email', email);
            body.set('username', username);
            body.set('password', password);
        }
        else if(this.props.dialog === "login"){
            
            if(username === "" || password === ""){
                return alert("Please Enter All Required Fields!");
            }
            body.set('username', username);
            body.set('password', password);
        }
        
        request({
            "method": 'POST', 
            "url": apiBasePath + endpoints[this.props.dialog], 
            "data": body,
            "credentials": 'same-origin',
            "withCredentials" : true,
            "config": { "headers": { "Content-Type": 'multipart/form-data' } } 
        })
        .then((resp) => {
            this.setState({
                "dialogOpen": false,
            });
            this.props.loginSuccess(resp)
        })
        .catch((err) => {
            if(err.response.status === 401){
                return alert(err.response.data.error)
            }
            else{
                return console.log(err)
            }
        });
    }

    render() {
        if(this.props.dialog === "register"){
            
            const actions = [
                <FlatButton 
                label="Cancel" 
                onClick={() => this.setState({ "dialogOpen": false })} />,

                <FlatButton 
                label="Register" 
                primary={true} 
                onClick={() => this.handleSubmit()} />
            ];
            return <Dialog 
                  title="Register"
                  actions={actions}
                  modal={false}
                  open={this.state.dialogOpen}
                  onRequestClose={() => this.setState({"dialogOpen": false})} >
                  <TextField
                      autoFocus
                      margin="normal"
                      id="name"
                      label="Name"
                      floatingLabelText="Full Name*"
                      type="text"
                      onChange={(e, val) => this.setState({ "name": val })}
                      fullWidth
                  />
                  
                  <TextField
                      margin="normal"
                      id="username"
                      label="Username"
                      floatingLabelText="Username*"
                      type="text"
                      onChange={(e, val) => this.setState({ "username": val })}
                      fullWidth
                  />
                  
                  <TextField
                      margin="normal"
                      id="email"
                      label="Email"
                      floatingLabelText="Email"
                      type="email"
                      onChange={(e, val) => this.setState({ "email": val })}
                      fullWidth
                  />
                  
                  <TextField
                      margin="normal"
                      id="password"
                      label="Password"
                      floatingLabelText="Password*"
                      type="password"
                      onChange={(e, val) => this.setState({ "password": val })}
                      fullWidth
                  />
                  
                  <TextField
                      margin="normal"
                      id="confirm"
                      label="confirm Password"
                      floatingLabelText="Confirm Password*"
                      type="password"
                      onChange={(e, val) => this.setState({ "confirm": val })}
                      fullWidth
                  />

              </Dialog>
        }
        else if(this.props.dialog === "login") {
            const actions = [
                <FlatButton 
                label="Cancel" 
                onClick={() => this.setState({ "dialogOpen": false })} />,

                <FlatButton 
                label="Login" 
                primary={true} 
                onClick={() => this.handleSubmit()} />
            ];
            return <Dialog 
                title="Login"
                actions={actions}
                modal={false}
                open={this.state.dialogOpen}
                onRequestClose={() => this.setState({ "dialogOpen": false })} >
                
                    <TextField
                        autoFocus
                        margin="normal"
                        id="username"
                        label="Username"
                        floatingLabelText="Username*"
                        type="text"
                        onChange={(e, val) => this.setState({ "username": val })}
                        fullWidth
                    />
                    <br/>
                    <TextField
                        margin="normal"
                        id="password"
                        label="Password"
                        floatingLabelText="Password*"
                        type="password"
                        onChange={(e, val) => this.setState({ "password": val })}
                        fullWidth
                    />

                </Dialog>
        }
        else {
            return <div/>
        }
    }
}

export default LoginRegister;