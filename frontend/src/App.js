import React, { useState } from 'react';
import axios from 'axios';
import ImageUpload from './components/ImageUpload';
import ResultsDisplay from './components/ResultsDisplay';
import CareInstructions from './components/CareInstructions';
import ErrorMessage from './components/ErrorMessage';
import Loading from './components/Loading';
import './App.css';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);

  const handleImageSelect = async (file) => {
    if (!file) {
      // Reset when user clears the image
      setResult(null);
      setError(null);
      return;
    }

    setError(null);
    setResult(null);
    setIsLoading(true);

    try {
      // Create form data for multipart upload
      const formData = new FormData();
      formData.append('image', file);

      // Call backend API
      const response = await axios.post(`${API_URL}/identify`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        timeout: 30000, // 30 second timeout
      });

      setResult(response.data);
    } catch (err) {
      console.error('Error identifying plant:', err);

      if (err.code === 'ECONNABORTED') {
        setError('Request timed out. Please try again.');
      } else if (err.response) {
        // Backend returned an error
        const errorMsg = err.response.data?.error || 'Failed to identify plant';
        setError(errorMsg);
      } else if (err.request) {
        // No response received
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
  };

  return (
    <div className="App">
      <header className="app-header">
        <h1 className="app-title">Succulent Identifier</h1>
        <p className="app-subtitle">Upload a photo to identify your succulent and get care instructions</p>
      </header>

      <main className="app-main">
        <ImageUpload
          onImageSelect={handleImageSelect}
          isLoading={isLoading}
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
            <ResultsDisplay plant={result.plant} />
            <CareInstructions care={result.care} />
          </>
        )}
      </main>

      <footer className="app-footer">
        <p>Powered by EfficientNet-B0 and PyTorch</p>
      </footer>
    </div>
  );
}

export default App;
