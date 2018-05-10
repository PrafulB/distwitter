import React, { Component } from 'react';
import {
  BrowserRouter as Router,
  Route,
} from 'react-router-dom';
import './App.css';
import Header from '../header'
import Login from '../login'
import Home from '../home'
import 'normalize-css'

import request from 'axios'

const endpoints = require("../../endpoints.json")

class App extends Component {

  constructor(props) {
    super(props);
    this.state = {
      "isLoggedIn" : false
    }
  }

  componentWillMount() {
    console.log("in componentWillMount");
    request(endpoints.apiBasePath + endpoints.checkLogin, {"withCredentials" : true})
    .then(user => { 
      console.log("true")
      this.setState({isLoggedIn : true})
    })
    .catch((err) => { 
      console.log(err)
      console.log(err.response.status)
    })
  } 

  handleLogout() {}

  handleOnAuth() {}

  render() {
    return (
      <Router>
        <div>
          <Header />
          <Route exact path='/' render={() => {
            if (this.state.isLoggedIn) {
              return (
                <Home />
              )
            } else {
              return ( <Login /> )
            }
          }} />
        </div>
      </Router>
    );
  }
}

export default App;
