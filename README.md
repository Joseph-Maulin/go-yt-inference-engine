# go-yt-inference-engine

Connect to youtube streams and run onnxruntime inference

## Requirements

### System Dependencies

- **FFmpeg** - For video stream processing and UDP broadcasting
  ```bash
  # macOS
  brew install ffmpeg
  
  # Ubuntu/Debian
  sudo apt-get install ffmpeg
  ```

- **OpenCV 4.x** - For video frame processing (required by GoCV)
  ```bash
  # macOS
  brew install opencv
  
  # Ubuntu/Debian
  sudo apt-get install libopencv-dev
  ```

- **uv** - Fast Python package installer and resolver
  ```bash
  # macOS/Linux
  curl -LsSf https://astral.sh/uv/install.sh | sh
  
  # Or with pip
  pip install uv
  ```

### Language Requirements

- **Go 1.25.1+** - Main application runtime
  - Download from [golang.org](https://golang.org/dl/)

- **Python 3.14+** - For yt-dlp stream extraction
  - Download from [python.org](https://www.python.org/downloads/)
  - Or use [uv](https://docs.astral.sh/uv/) for dependency management

## Installation

1. Install system dependencies (FFmpeg, OpenCV)
2. Clone the repository
3. Install Go dependencies:
   ```bash
   go mod download
   ```
4. Install Python dependencies:
   ```bash
   uv sync
   ```

## Quick Start

```bash
# Run tests
go test ./... -v

# Run broadcast test
go test ./pkg/services/broadcast/... -v
```
