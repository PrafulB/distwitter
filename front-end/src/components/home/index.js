import React, {Component} from 'react';
// import {
//   BrowserRouter as Router,
//   Route,
//   Link
// } from 'react-router-dom';
import './style.css'
// import moment from 'moment'
// import messages from './messages.json'

import HomeCardsList from '../homeCardsList'
import ProfileBar from '../profileBar'
import InputText from '../inputText'
import * as request from 'axios';

const endpoints = require("../../endpoints.json")

class Home extends Component {
	constructor(props){
        super(props)
        this.state = {
            messages : [],
            openText : false,
            user: {},
        }

        this.handleOpenText = this.handleOpenText.bind(this)
    }

    componentWillMount() {
        request(endpoints.apiBasePath + endpoints.checkLogin, {"withCredentials" : true})
        .then(resp => { 
            this.setState({
                user: resp.data.user
            })
            return request({
                "method": 'GET', 
                "url": endpoints.apiBasePath + endpoints.getPosts, 
                "credentials": 'same-origin',
                "withCredentials" : true,
                "config": { "headers": { 'Content-Type': 'application/x-www-form-urlencoded' } } 
            })
        })
        .then((resp) => {
            this.setState({messages : resp.data.posts})
        })
        .catch((err) => { 
            console.log(err)
        })
    }

    renderOpenText () {
        if (this.state.openText) {
            return (
                <InputText
                  onCloseText={this.handleCloseText.bind(this)}
                  post={this.sendTweet.bind(this)}
                />
            )
        }
    }

    handleOpenText (event) {
        event.preventDefault()
        this.setState({ openText: true })
    }

    handleCloseText (event) {
        event.preventDefault()
        this.setState({ openText: false })
    }

    logout() {
        request({
            method: "GET",
            url: endpoints.apiBasePath + endpoints.logout,
            "credentials": 'same-origin',
            "withCredentials" : true,
            "config": { "headers": { 'Content-Type': 'application/x-www-form-urlencoded' } } 
        })
        .then((resp) => {
            window.location.reload()
        })
    }

    sendTweet(text){
        let body = new FormData()
        body.set("status", text)
        request({
            method: 'POST',
            url: endpoints.apiBasePath + endpoints.post,
            data: body,
            "credentials": 'same-origin',
            "withCredentials" : true,
            "config": { "headers": { 'Content-Type': 'application/x-www-form-urlencoded' } } 
        })
        this.setState({
            openText: false
        })
        window.location.reload()
    }

    deleteAccount(){
        request({
            method: 'DELETE',
            url: endpoints.apiBasePath + endpoints.delete,
            "credentials": 'same-origin',
            "withCredentials" : true,
            "config": { "headers": { 'Content-Type': 'application/x-www-form-urlencoded' } } 
        })
        window.location.reload()
    }

    render() {
    	return(
    		<div>
                <ProfileBar
                  picture=""
                  username={this.state.user.Username}
                  onOpenText={this.handleOpenText}
                  onLogout={this.logout}
                  onDelete={this.deleteAccount}
                />
                {this.renderOpenText()}
				<HomeCardsList
					messages={this.state.messages}
					onRetweet={this.handleRetweet}
					onFavorite={this.handleFavorite}
					onReplyTweet={this.handleReplyTweet}
				/>
			</div>
    	)
    }
}

export default Home;