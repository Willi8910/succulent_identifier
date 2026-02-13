import React from 'react';
import './Loading.css';

const Loading = () => {
  return (
    <div className="loading-container">
      <div className="spinner"></div>
      <p className="loading-text">Identifying your succulent...</p>
      <p className="loading-subtext">This may take a few seconds</p>
    </div>
  );
};

export default Loading;
