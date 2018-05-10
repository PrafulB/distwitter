import React, { PropTypes } from 'react';
import {
  BrowserRouter as Router,
  Route,
  Link
} from 'react-router-dom';
import './style.css'
import moment from 'moment'
import HomeCards from '../homeCards'

// const propTypes = {
//   messages: PropTypes.arrayOf(PropTypes.Object).isRequired,
//   onRetweet: PropTypes.func.isRequired,
//   onFavorite: PropTypes.func.isRequired,
//   onReplyTweet: PropTypes.func.isRequired
// }

function HomeCardsList ({ messages, onRetweet, onFavorite, onReplyTweet }) {
  return (
    <div className="rootHomeCardList">
      {messages.map(msg => {
        return (
          <HomeCards
            key={msg.PostId}
            text={msg.Content}
            picture={msg.picture}
            username={msg.Username}
            date={msg.timePosted}
            numRetweets={msg.retweets}
            numFavorites={msg.favorites}
            onRetweet={() => onRetweet(msg.id)}
            onFavorite={() => onFavorite(msg.id)}
            onReplyTweet={() => onReplyTweet(msg.id, msg.username)}
          />
        )
      }).reverse()}
    </div>
  )
}

// HomeCardsList.propTypes = propTypes

export default HomeCardsList;