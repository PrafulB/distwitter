import React, {Component} from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import RaisedButton from 'material-ui/RaisedButton';
import LoginRegister from './loginregister';

const style = {
  margin: 12,
}

class Login extends Component {
  constructor(props){
    super(props);
    this.state = {
      modalOpen: '',
    };
  }

  openDialog(modal){
    this.setState({
      modalOpen: modal
    });
  }
  
  loginSuccess(resp){
    if(resp.status === 200) {
      window.location.href = "http://localhost:3000/"
    }
  }

  render() {
    return (
      <div className="App">
        <MuiThemeProvider>
          <div>
            <RaisedButton label="Register" style={style} primary={true} onClick={() => {this.openDialog('register')}} />
            <RaisedButton label="Login" style={style} primary={true} onClick={() => {this.openDialog('login')}} />
            <LoginRegister 
              dialog={this.state.modalOpen}
              loginSuccess={this.loginSuccess}
            />
          </div>
        </MuiThemeProvider>
      </div>
    );
  }
}

export default Login;