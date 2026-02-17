import React, { useState } from 'react';
import ImageUpload from './components/ImageUpload';
import ResultsDisplay from './components/ResultsDisplay';
import CareInstructions from './components/CareInstructions';
import ErrorMessage from './components/ErrorMessage';
import Loading from './components/Loading';
import ChatDrawer from './components/ChatDrawer';
import HistorySidebar from './components/HistorySidebar';
import { identifyPlant, getHistoryById } from './services/api';
import './App.css';

function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);
  const [isChatDrawerOpen, setIsChatDrawerOpen] = useState(false);
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleImageSelect = async (file) => {
    if (!file) {
      // Reset when user clears the image
      setResult(null);
      setError(null);
      setIsChatDrawerOpen(false);
      return;
    }

    setError(null);
    setResult(null);
    setIsChatDrawerOpen(false);
    setIsLoading(true);

    try {
      const data = await identifyPlant(file);
      setResult(data);
      // Trigger history sidebar refresh
      setRefreshTrigger(prev => prev + 1);
    } catch (err) {
      console.error('Error identifying plant:', err);

      if (err.code === 'ECONNABORTED') {
        setError('Request timed out. Please try again.');
      } else if (err.response) {
        const errorMsg = err.response.data?.error || 'Failed to identify plant';
        setError(errorMsg);
      } else if (err.request) {
        setError('Unable to connect to the server. Please make sure the backend is running.');
      } else {
        setError('An unexpected error occurred. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRetry = () => {
    setError(null);
    setResult(null);
    setIsChatDrawerOpen(false);
  };

  const handleOpenChat = () => {
    setIsChatDrawerOpen(true);
  };

  const handleCloseChat = () => {
    setIsChatDrawerOpen(false);
  };

  const handleToggleSidebar = () => {
    setIsSidebarOpen(!isSidebarOpen);
  };

  const handleSelectHistory = async (historyItem) => {
    setIsLoading(true);
    setError(null);
    setIsChatDrawerOpen(false);

    try {
      const data = await getHistoryById(historyItem.id);

      // Construct image URL
      const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
      const imageUrl = data.image_path ? `${API_URL}/uploads/${data.image_path}` : null;

      // Transform the data to match the expected result format
      const transformedResult = {
        id: data.id,
        plant: {
          genus: data.genus,
          species: data.species || '',
          confidence: data.confidence
        },
        care: data.care_guide || {
          sunlight: 'Information not available',
          watering: 'Information not available',
          soil: 'Information not available',
          notes: ''
        },
        imageUrl: imageUrl
      };

      setResult(transformedResult);

      // Close sidebar on mobile after selection
      if (window.innerWidth <= 768) {
        setIsSidebarOpen(false);
      }
    } catch (err) {
      console.error('Failed to load history item:', err);
      setError('Failed to load this identification. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="App">
      <HistorySidebar
        isOpen={isSidebarOpen}
        onToggle={handleToggleSidebar}
        onSelectHistory={handleSelectHistory}
        currentId={result?.id}
        refreshTrigger={refreshTrigger}
      />

      <div className={`app-content ${isSidebarOpen ? 'sidebar-open' : ''}`}>
        <header className="app-header">
          <h1 className="app-title">Succulent Identifier</h1>
          <p className="app-subtitle">Upload a photo to identify your succulent and get care instructions</p>
        </header>

        <main className="app-main">
        <ImageUpload
          onImageSelect={handleImageSelect}
          isLoading={isLoading}
          externalImageUrl={result?.imageUrl}
        />

        {isLoading && <Loading />}

        {error && (
          <ErrorMessage
            message={error}
            onRetry={handleRetry}
          />
        )}

        {result && !isLoading && (
          <>
            <ResultsDisplay plant={result.plant} onOpenChat={handleOpenChat} />
            <CareInstructions care={result.care} />
          </>
        )}
        </main>

        <footer className="app-footer">
          <p>Created by <a href="https://github.com/Willi8910" target="_blank" rel="noopener noreferrer">William</a>, Goose farmer wannabe</p>
          <p className="footer-links">
            <a href="https://github.com/Willi8910" target="_blank" rel="noopener noreferrer">GitHub</a>
            {' • '}
            <a href="https://www.linkedin.com/in/williamlie8910/" target="_blank" rel="noopener noreferrer">LinkedIn</a>
          </p>
        </footer>
      </div>

      <ChatDrawer
        isOpen={isChatDrawerOpen}
        onClose={handleCloseChat}
        identification={result}
      />
    </div>
  );
}

export default App;
