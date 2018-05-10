import React, {Component} from 'react';
import {
  BrowserRouter as Router,
  Route,
  Link
} from 'react-router-dom';
import './style.css'
import moment from 'moment'

class HomeCards extends Component {
	constructor(props){
        super(props)
        this.state = {
            
        }
    }

    render() {
    	let dateFormat = moment(this.props.date).fromNow();
    	let userLink = `/user/${this.props.username}`
    	
    	return(
    		<div className="root">
				<div className="user">
					<span className="displayName">{this.props.displayName}</span>
					<span className="username"> {this.props.username}</span>
					<span className="date">{dateFormat}</span>
				 </div>
				
				<h3>{this.props.text}</h3>
				
				<div className="buttons">
					<div className="icon" onClick={this.props.onReplyTweet}>
						<span className='fa fa-reply'></span>
					</div>
					
					<div className={(this.state.pressRetweet) ? "rtGreen" : "space"} onClick={this.onPressRetweet}>
						<span className='fa fa-retweet'></span>
						<span className="number">{this.props.numRetweets}</span>
					</div>
					<div className={(this.state.pressFavorite) ? "favYellow" : ''} onClick={this.onPressFavorite} >
						<span className='fa fa-star'></span>
						<span className="number">{this.props.numFavorites}</span>
					</div>
				</div>
			</div>
    	)
    }
}

export default HomeCards;