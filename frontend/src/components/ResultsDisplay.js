import React from 'react';
import './ResultsDisplay.css';

const ResultsDisplay = ({ plant }) => {
  if (!plant) return null;

  const { genus, species, confidence } = plant;
  const confidencePercent = (confidence * 100).toFixed(1);
  const isHighConfidence = confidence >= 0.4;

  return (
    <div className="results-container">
      <div className="results-header">
        <h2 className="results-title">Identification Results</h2>
      </div>

      <div className="plant-info">
        <div className="plant-name">
          <div className="label">Genus</div>
          <div className="value genus">{genus}</div>
        </div>

        {species && (
          <div className="plant-name">
            <div className="label">Species</div>
            <div className="value species">{species}</div>
          </div>
        )}

        {!species && (
          <div className="low-confidence-message">
            <p>Species uncertain - showing genus-level information</p>
          </div>
        )}

        <div className="confidence-section">
          <div className="label">Confidence</div>
          <div className="confidence-bar-container">
            <div
              className={`confidence-bar ${isHighConfidence ? 'high' : 'low'}`}
              style={{ width: `${confidencePercent}%` }}
            ></div>
          </div>
          <div className="confidence-text">{confidencePercent}%</div>
        </div>
      </div>
    </div>
  );
};

export default ResultsDisplay;
