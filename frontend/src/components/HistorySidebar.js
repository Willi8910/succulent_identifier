import React, { useState, useEffect } from 'react';
import { getHistory } from '../services/api';
import './HistorySidebar.css';

const HistorySidebar = ({ isOpen, onToggle, onSelectHistory, currentId, refreshTrigger }) => {
  const [historyItems, setHistoryItems] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isOpen) {
      loadHistory();
    }
  }, [isOpen, refreshTrigger]);

  const loadHistory = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await getHistory(50, 0); // Load last 50 items
      setHistoryItems(response.items || []);
    } catch (err) {
      console.error('Failed to load history:', err);
      setError('Failed to load history');
    } finally {
      setIsLoading(false);
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMs = now - date;
    const diffInHours = diffInMs / (1000 * 60 * 60);

    if (diffInHours < 24) {
      return date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
      });
    } else if (diffInHours < 48) {
      return 'Yesterday';
    } else {
      return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric'
      });
    }
  };

  return (
    <>
      {/* Mobile Toggle Button */}
      <button className="sidebar-toggle" onClick={onToggle}>
        {isOpen ? '✕' : '📋'}
      </button>

      {/* Sidebar */}
      <div className={`history-sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-header">
          <h2>History</h2>
          <button className="refresh-button" onClick={loadHistory} disabled={isLoading}>
            🔄
          </button>
        </div>

        <div className="sidebar-content">
          {isLoading && (
            <div className="sidebar-loading">
              <div className="loading-spinner"></div>
              <p>Loading history...</p>
            </div>
          )}

          {error && (
            <div className="sidebar-error">
              <p>{error}</p>
              <button onClick={loadHistory}>Retry</button>
            </div>
          )}

          {!isLoading && !error && historyItems.length === 0 && (
            <div className="sidebar-empty">
              <p>No history yet</p>
              <p className="sidebar-empty-subtitle">
                Identify your first succulent to get started!
              </p>
            </div>
          )}

          {!isLoading && !error && historyItems.length > 0 && (
            <div className="history-list">
              {historyItems.map((item) => (
                <div
                  key={item.id}
                  className={`history-item ${item.id === currentId ? 'active' : ''}`}
                  onClick={() => onSelectHistory(item)}
                >
                  <div className="history-item-content">
                    <div className="history-item-name">
                      <span className="genus">{item.genus}</span>
                      {item.species && (
                        <span className="species">{item.species}</span>
                      )}
                    </div>
                    <div className="history-item-meta">
                      <span className="confidence">
                        {(item.confidence * 100).toFixed(0)}%
                      </span>
                      <span className="date">{formatDate(item.created_at)}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Overlay for mobile */}
      {isOpen && (
        <div className="sidebar-overlay" onClick={onToggle}></div>
      )}
    </>
  );
};

export default HistorySidebar;
