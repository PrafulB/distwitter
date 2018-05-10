import React, { PropTypes } from 'react'
import { Link } from 'react-router-dom'
import './style.css'

// const propTypes = {
//   picture: PropTypes.string.isRequired,
//   username: PropTypes.string.isRequired,
//   onOpenText: PropTypes.func.isRequired
// }

function ProfileBar ({ picture, username, onOpenText, onLogout, onDelete }) {
  return (
    <div className="rootProfileBar">
      <Link to='/profile'>
        <span className="usernameProfileBar"> Hello @{username}!</span>
      </Link>
      <button onClick={onOpenText} className="buttonProfileBar">
        <span className='fa fa-lg fa-edit'></span> Tweet!
      </button>
      <button onClick={onLogout} className="buttonProfileBar">
        <span className='fa fa-sign-out'></span> Logout
      </button>
      <button style={{position: "absolute", right: 15}} onClick={onDelete} className="buttonProfileBar">
        <span className='fa fa-delete'></span> Delete My Account
      </button>
    </div>
  )
}

// ProfileBar.propTypes = propTypes

export default ProfileBar