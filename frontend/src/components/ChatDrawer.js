import React, { useState, useEffect, useRef } from 'react';
import { sendChatMessage, getChatHistory } from '../services/api';
import './ChatDrawer.css';

const ChatDrawer = ({ isOpen, onClose, identification }) => {
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Load chat history when drawer opens
  useEffect(() => {
    if (isOpen && identification?.id) {
      loadChatHistory();
    }
  }, [isOpen, identification]);

  const loadChatHistory = async () => {
    try {
      const response = await getChatHistory(identification.id);
      setMessages(response.messages || []);
    } catch (err) {
      console.error('Failed to load chat history:', err);
      // Don't show error for empty history
      setMessages([]);
    }
  };

  const handleSendMessage = async (e) => {
    e.preventDefault();

    if (!inputMessage.trim() || isLoading) return;

    const userMessage = {
      message: inputMessage,
      sender: 'user',
      created_at: new Date().toISOString(),
    };

    // Add user message immediately
    setMessages(prev => [...prev, userMessage]);
    setInputMessage('');
    setIsLoading(true);
    setError(null);

    try {
      const response = await sendChatMessage(identification.id, inputMessage);

      // Add LLM response
      const llmMessage = {
        id: response.message_id,
        message: response.message,
        sender: 'llm',
        created_at: response.timestamp,
      };

      setMessages(prev => [...prev, llmMessage]);
    } catch (err) {
      console.error('Failed to send message:', err);
      setError('Failed to get response. Please try again.');

      // Remove the user message that failed
      setMessages(prev => prev.slice(0, -1));
    } finally {
      setIsLoading(false);
    }
  };

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (!identification) return null;

  return (
    <>
      {/* Overlay */}
      <div
        className={`drawer-overlay ${isOpen ? 'open' : ''}`}
        onClick={onClose}
      />

      {/* Drawer */}
      <div className={`chat-drawer ${isOpen ? 'open' : ''}`}>
        {/* Header */}
        <div className="chat-header">
          <div className="chat-header-content">
            <h3>Chat about your plant</h3>
            <p className="chat-plant-name">
              {identification.plant.genus}
              {identification.plant.species && ` ${identification.plant.species}`}
            </p>
          </div>
          <button className="close-button" onClick={onClose}>
            ✕
          </button>
        </div>

        {/* Messages */}
        <div className="chat-messages">
          {messages.length === 0 && !isLoading && (
            <div className="chat-welcome">
              <p>👋 Ask me anything about your {identification.plant.genus}!</p>
              <p className="chat-welcome-subtext">
                I can help with watering, sunlight, propagation, and more.
              </p>
            </div>
          )}

          {messages.map((msg, index) => (
            <div
              key={msg.id || index}
              className={`message ${msg.sender === 'user' ? 'user' : 'llm'}`}
            >
              <div className="message-content">
                {msg.message}
              </div>
              {msg.created_at && (
                <div className="message-timestamp">
                  {formatTimestamp(msg.created_at)}
                </div>
              )}
            </div>
          ))}

          {isLoading && (
            <div className="message llm">
              <div className="message-content typing">
                <span></span>
                <span></span>
                <span></span>
              </div>
            </div>
          )}

          {error && (
            <div className="chat-error">
              {error}
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <form className="chat-input-container" onSubmit={handleSendMessage}>
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            placeholder="Ask about care, watering, sunlight..."
            className="chat-input"
            disabled={isLoading}
          />
          <button
            type="submit"
            className="send-button"
            disabled={isLoading || !inputMessage.trim()}
          >
            Send
          </button>
        </form>
      </div>
    </>
  );
};

export default ChatDrawer;
