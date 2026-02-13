# Succulent Identifier - Frontend

React-based web interface for the Succulent Identifier application. Upload a photo of your succulent plant to identify the species and receive personalized care instructions.

## Features

- **Drag-and-Drop Upload**: Intuitive image upload with drag-and-drop support
- **Real-time Validation**: Client-side validation for file type (JPG/PNG) and size (max 5MB)
- **Image Preview**: See your uploaded image before identification
- **Confidence Visualization**: Visual confidence bar showing prediction certainty
- **Care Instructions**: Comprehensive care guidance (sunlight, watering, soil)
- **Error Handling**: User-friendly error messages with retry functionality
- **Loading States**: Clear feedback during image processing
- **Responsive Design**: Optimized for desktop, tablet, and mobile devices
- **Modern UI**: Clean, gradient-based design with smooth animations

## Technology Stack

- **React 18**: Modern React with hooks
- **Axios**: HTTP client for API communication
- **CSS3**: Custom styling with gradients and animations
- **Create React App**: Build tooling and development server

## Project Structure

```
frontend/
├── public/              # Static assets
├── src/
│   ├── components/      # React components
│   │   ├── ImageUpload.js         # Drag-and-drop upload
│   │   ├── ImageUpload.css
│   │   ├── ResultsDisplay.js      # Plant identification results
│   │   ├── ResultsDisplay.css
│   │   ├── CareInstructions.js    # Plant care guide
│   │   ├── CareInstructions.css
│   │   ├── ErrorMessage.js        # Error handling UI
│   │   ├── ErrorMessage.css
│   │   ├── Loading.js             # Loading spinner
│   │   └── Loading.css
│   ├── App.js           # Main application component
│   ├── App.css          # App-level styles
│   ├── index.js         # Entry point
│   ├── index.css        # Global styles
│   └── ...
├── package.json
└── README.md
```

## Prerequisites

- **Node.js**: >= 14.0.0
- **npm**: >= 6.0.0
- **Backend API**: Running on `http://localhost:8080` (see `../backend/README.md`)

## Installation

1. **Navigate to frontend directory**:
   ```bash
   cd frontend
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

## Configuration

The frontend connects to the backend API using an environment variable:

**Default**: `http://localhost:8080`

To override, create a `.env` file:

```bash
REACT_APP_API_URL=http://your-backend-url:port
```

## Running the Application

### Development Mode

Start the development server with hot reload:

```bash
npm start
```

The app will open at `http://localhost:3000`.

### Production Build

Build the app for production:

```bash
npm run build
```

Optimized files will be in the `build/` directory.

### Serve Production Build

To test the production build locally:

```bash
npm install -g serve
serve -s build
```

## Usage

1. **Start the backend services**:
   - ML Service: Running on port 8000
   - Backend API: Running on port 8080

2. **Start the frontend**:
   ```bash
   npm start
   ```

3. **Upload an image**:
   - Drag and drop an image onto the upload area, OR
   - Click the upload area to browse for a file
   - Supported formats: JPG, PNG
   - Maximum size: 5MB

4. **View results**:
   - Genus and species identification
   - Confidence percentage with visual bar
   - Comprehensive care instructions

## Components

### ImageUpload

Handles file upload with validation:
- Drag-and-drop interface
- File type validation (JPG/PNG only)
- File size validation (max 5MB)
- Image preview
- Reset functionality

**Props**:
- `onImageSelect(file)`: Callback when image is selected
- `isLoading`: Boolean to disable during processing

### ResultsDisplay

Shows plant identification results:
- Genus name (always shown)
- Species name (shown if confidence >= 0.4)
- Confidence percentage with color-coded bar
- Low confidence warning message

**Props**:
- `plant`: Object with `{ genus, species, confidence }`

### CareInstructions

Displays plant care guidance:
- Sunlight requirements
- Watering schedule
- Soil recommendations
- Additional notes

**Props**:
- `care`: Object with `{ sunlight, watering, soil, notes }`

### ErrorMessage

User-friendly error display:
- Error icon and message
- Optional retry button

**Props**:
- `message`: Error message string
- `onRetry`: Callback for retry button (optional)

### Loading

Animated loading indicator:
- Spinning loader animation
- Informative text

## API Integration

The frontend communicates with the backend via REST API:

**Endpoint**: `POST /identify`

**Request**:
- Content-Type: `multipart/form-data`
- Body: `image` field with file data

**Response**:
```json
{
  "plant": {
    "genus": "Haworthia",
    "species": "Haworthia Zebrina",
    "confidence": 0.9468
  },
  "care": {
    "sunlight": "Bright, indirect light...",
    "watering": "Water when soil is dry...",
    "soil": "Well-draining cactus mix...",
    "notes": "Optional care notes"
  }
}
```

## Error Handling

The app handles various error scenarios:

- **Network errors**: "Unable to connect to the server..."
- **Timeout**: "Request timed out. Please try again."
- **Backend errors**: Shows error message from API response
- **Invalid file type**: Client-side alert
- **File too large**: Client-side alert

## Responsive Design

Breakpoints:
- **Desktop**: > 768px (full features)
- **Tablet**: 600px - 768px (medium sizing)
- **Mobile**: < 600px (optimized layout)

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Performance

- **Initial load**: < 3 seconds
- **API request**: < 2 seconds (depends on backend)
- **Image preview**: Instant
- **Hot reload**: Enabled in development

## Styling

Custom CSS with:
- CSS3 gradients
- Flexbox layouts
- CSS animations
- Media queries for responsiveness
- Box-shadow for depth
- Smooth transitions

No CSS frameworks (pure CSS for minimal bundle size).

## Development

### Available Scripts

- `npm start`: Development server
- `npm run build`: Production build
- `npm test`: Run tests (not yet implemented)
- `npm run eject`: Eject from Create React App (⚠️ irreversible)

### Adding New Components

1. Create component file in `src/components/`
2. Create corresponding CSS file
3. Import and use in `App.js`

### Code Style

- Functional components with hooks
- Props destructuring
- Clear component naming
- Separate CSS files per component
- Mobile-first responsive design

## Known Issues

- None currently

## Future Enhancements

- [ ] Add image cropping before upload
- [ ] Support for multiple image uploads
- [ ] History of previous identifications
- [ ] Share results functionality
- [ ] PWA support for offline capability
- [ ] Unit tests with Jest and React Testing Library
- [ ] E2E tests with Cypress
- [ ] Internationalization (i18n)

## Troubleshooting

**Issue**: Cannot connect to backend
- **Solution**: Ensure backend is running on `http://localhost:8080`
- Check `REACT_APP_API_URL` environment variable

**Issue**: File upload fails
- **Solution**: Check file type (JPG/PNG) and size (< 5MB)
- Verify backend is accepting requests

**Issue**: Blank screen
- **Solution**: Check browser console for errors
- Ensure all dependencies are installed: `npm install`

**Issue**: Compilation errors
- **Solution**: Clear cache and reinstall:
  ```bash
  rm -rf node_modules package-lock.json
  npm install
  ```

## Contributing

When adding features:
1. Follow existing component structure
2. Add responsive CSS
3. Handle loading and error states
4. Test on mobile devices
5. Update this README

## License

Part of the Succulent Identifier project.

## Related Documentation

- Backend API: `../backend/README.md`
- ML Service: `../ml_service/README.md`
- Project Overview: `../README.md` (to be created)
- Technical Design: `../TDD.txt`
- Product Requirements: `../PRD.txt`

---

**Frontend Development Complete** ✅

For support or questions, refer to the main project documentation.
