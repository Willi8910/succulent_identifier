import React from 'react';
import './CareInstructions.css';

const CareInstructions = ({ care }) => {
  if (!care) return null;

  const { sunlight, watering, soil, notes, trivia } = care;

  return (
    <div className="care-container">
      <div className="care-header">
        <h2 className="care-title">Care Instructions</h2>
      </div>

      <div className="care-content">
        {trivia && (
          <div className="care-trivia">
            <div className="care-icon">💡</div>
            <div className="care-details">
              <div className="care-label">Trivia</div>
              <div className="care-text">{trivia}</div>
            </div>
          </div>
        )}

        <div className="care-item">
          <div className="care-icon">☀️</div>
          <div className="care-details">
            <div className="care-label">Sunlight</div>
            <div className="care-text">{sunlight}</div>
          </div>
        </div>

        <div className="care-item">
          <div className="care-icon">💧</div>
          <div className="care-details">
            <div className="care-label">Watering</div>
            <div className="care-text">{watering}</div>
          </div>
        </div>

        <div className="care-item">
          <div className="care-icon">🌱</div>
          <div className="care-details">
            <div className="care-label">Soil</div>
            <div className="care-text">{soil}</div>
          </div>
        </div>

        {notes && (
          <div className="care-notes">
            <div className="care-icon">📝</div>
            <div className="care-details">
              <div className="care-label">Notes</div>
              <div className="care-text">{notes}</div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default CareInstructions;
