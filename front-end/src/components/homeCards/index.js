import React, {Component} from 'react';
import {
  BrowserRouter as Router,
  Route,
  Link
} from 'react-router-dom';
import './style.css'
import moment from 'moment'
import request from 'axios'

const endpoints = require('../../endpoints.json')

class HomeCards extends Component {
	constructor(props){
        super(props)
        this.state = {
            user:{}
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
		  this.setState({
			  "user": resp.data.user,
		  });
	  })
	  .catch((err) => {
		  if(err.response.status === 409){
			  return alert(err.response.data.error)
		  }
		  else{
			  return console.log(err)
		  }
	  });
	  }
	
	followUser(username){
		if(username != this.state.user.Username && this.state.user.Following.indexOf(username) == -1){
			console.log("FOLLOWING")
			let body = new FormData()
			body.set("username", username)
			request({
				method: "POST",
				url: endpoints.apiBasePath + endpoints.follow,
				data: body,
				"credentials": 'same-origin',
				"withCredentials" : true,
				"config": { "headers": { "Content-Type": 'multipart/form-data' } } 
			})
			.then(()=> window.location.reload())

		}
	}

    render() {
    	let dateFormat = moment(this.props.date).fromNow();
    	// let userLink = `/user/${this.props.username}`
    	
    	return(
    		<div className="root">
				<div className="user">
					<b><u><span onClick={() => this.followUser(this.props.username)} className="username"> {this.props.username}</span></u></b>
					<span className="date">{dateFormat}</span>
				</div>
				
				<h3>{this.props.text}</h3>
				
				<div className="buttons">
					<div className="icon" >
						<span className='fa fa-reply'></span>
					</div>
					
					<div className={(this.state.pressRetweet) ? "rtGreen" : "space"} >
						<span className='fa fa-retweet'></span>
					</div>
				</div>
			</div>
    	)
    }
}

export default HomeCards;