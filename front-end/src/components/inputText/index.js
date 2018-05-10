import React, { Component } from 'react'
import styles from './style.css'
import request from 'axios'

const endpoints = require('../../endpoints.json')

// const propTypes = {
//   onSendText: PropTypes.func.isRequired,
//   onCloseText: PropTypes.func.isRequired,
//   userNameToReply: PropTypes.string.isRequired
// }

class InputText extends Component {
  
  constructor(props){
    super(props)
    this.state = {
      content: ""
    }
  }

  handleChange (event) {
    this.setState({
      content: event.target.value
    });
  }

  sendTweet(event) {
    event.preventDefault()
    this.props.post(this.state.content)
  }

  render(){
    return (
      <form onSubmit={this.sendTweet.bind(this)}>
        <textarea className="textTweet" name='text' maxLength={100} value={this.state.content} onChange={this.handleChange.bind(this)}/>
        <div className="buttonsTweet">
          <button className="closeTweet" onClick={this.props.onCloseText}>Close</button>
          <button className="sendTweet" type='submit'> Send </button>
        </div>
      </form>
    )
  }
}


// InputText.propTypes = propTypes

export default InputText