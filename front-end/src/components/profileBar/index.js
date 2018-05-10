import React, { Component } from 'react'
import { Link } from 'react-router-dom'
import './style.css'
import request from 'axios'

const endpoints = require('../../endpoints.json')
// const propTypes = {
//   picture: PropTypes.string.isRequired,
//   username: PropTypes.string.isRequired,
//   onOpenText: PropTypes.func.isRequired
// }

class ProfileBar extends Component {

  constructor(props){
    super(props)
    this.state = {
      following: []
    }
  }

  componentWillMount(){
    request({
      "method": 'POST', 
      "url": endpoints.apiBasePath + endpoints.checkLogin, 
      "credentials": 'same-origin',
      "withCredentials" : true,
      "config": { "headers": { "Content-Type": 'multipart/form-data' } } 
  })
  .then((resp) => {
    console.log(resp.data.user)
      this.setState({
          "following": resp.data.user.Following,
      });
  })
  .catch((err) => {
      console.log(err)
  });
  }
  render(){
    
      return (
        <div className="rootProfileBar">
          <Link to='/profile'>
            <span className="usernameProfileBar"> Hello @{this.props.username}!</span>
          </Link>
          <button onClick={this.props.onOpenText} className="buttonProfileBar">
            <span className='fa fa-lg fa-edit'></span> Tweet!
          </button>
          <button onClick={this.props.onLogout} className="buttonProfileBar">
            <span className='fa fa-sign-out'></span> Logout
          </button>
          <button style={{position: "absolute", right: 15}} onClick={this.props.onDelete} className="buttonProfileBar">
            <span className='fa fa-delete'></span> Delete My Account
          </button>
          <br/>
          <span className='usernameProfileBar'>Following: {this.state.following ? this.state.following.join() : "None"}</span> 
        </div>
      )
  }
}

// ProfileBar.propTypes = propTypes

export default ProfileBar