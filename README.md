# Web Analyzer

A web crawler application designed to scrape and analyze web page data, including HTML version, title, internal links, external links, and link status verification (active/inactive).

## Features

- **Page Analysis**: Extracts HTML version, page title, and categorizes links
- **Link Validation**: Checks internal and external links for availability
- **Performance Optimization**: Uses job queues with worker pools for efficient link checking
- **Real-time Streaming**: Streams data to frontend as it gets processed with http/1.1 server sent events
- **High Performance**: Optimized for handling large pages efficiently

## Running Instructions

### Backend Server (Port : 8000)

#### Using Docker Compose
```bash
docker-compose up
```

#### Using Go directly
```bash
cd cmd/server
go run main.go
```

### Frontend Client (port : 5173)

For the client application, navigate to the client repository ( https://github.com/RidmaTP/web-analyzer-fe ) and follow these instructions:

#### Using Docker Compose
```bash
docker-compose up
```

#### Using npm
```bash
npm i
npm run dev
```

## API Usage

Use the following curl command to interact with the API:

```bash
curl --location 'http://localhost:8000/api/result?url=' \
--data ''
```

## Challenges Faced and Solutions

### 1. Resource-Intensive Link Checking
**Challenge**: Checking each link for availability is highly resource-intensive and time-consuming.

**Solution**: Implemented a job queue with a worker pool to increase performance and handle concurrent link validation efficiently.

### 2. Large Page Processing Time
**Challenge**: Scanning very large pages takes significant time, leading to poor user experience.

**Solution**: Implemented real-time data streaming to send processed data to the frontend immediately as it becomes available, rather than waiting for complete processing.

## Future Improvements

1. **HTTP/2 Upgrade**: Replace HTTP 1.1 Server-Sent Events (SSE) with HTTP/2 for more efficient data streaming

2. **Monitoring Integration**: Add Prometheus integration for comprehensive metrics monitoring and observability

3. **Incremental Updates**: Implement change tracking to send only data changes to the frontend instead of transmitting the entire data object each time

4. **Enhanced Error Handling**: Improve error handling mechanisms across the application

5. **CORS Configuration**: Configure CORS to accept requests only from allowed origins for better security

6. **Test Coverage**: Add more unit tests to cover all scenarios and edge cases

7. **UI Enhancement**: Improve the user interface for better user experience and functionality
