# Research on Mux media processing

- Mux uses a distributed architecture to split the workload across multiple servers.
- The media is first analyzed using `ffprobe` to extract metadata, like the keyframes.
- The media is split into smaller chunks, around 2 to 6 seconds long.
- The information about these chunks is thrown into a distributed message queue (Kafka, RabbitMQ)
- The chunks are consumed by a pool of workers that process them in parallel using `ffmpeg`
- Each worker performs a specific tasks, like encoding, transcoding resolution, thumbnail generation, etc.
- Each chunk is uploaded to a storage service (S3)
- Instead of merging the chunks back into a single file, Mux generates an HLS manifest `.m3u8` file that references the individual chunks, and defines the playback order and timing.

### JIT Model

- Instead of pre-processing the entire variations for a video, Mux waits until the moment a user requests playback.
- It is only at this moment that Mux creates the HLS manifest or DASH packages.
- The result is then cached to CDN to avoid re-processing the same video for subsequent requests.

### Separation of Concerns Control vs Data

- Mux clearly separates the Go plane that handles request, webhooks, database operations, authentication, billing, etc; and the Data plane that handles the actual media manipulation and processing and its monitoring data.

### Machine Learning for Encoding Optimization

- During the probing phase, Mux dynamically adjusts the ffmpeg parameters based on the content of the video, to optimize the encoding process and reduce the file size without sacrificing quality.

### Direct to Cloud Uploads

- The user uploads media directly to the cloud storage (S3)
- The user first communicated with the Mux API with the intent, and Mux returns a presigned URL for the upload and a new video entry record in the database with a corresponding "pending" status.
- The presigned URL allows the user to upload the media directly to S3 without going through Mux servers.
- For most if not all uploads, Mux uses `UpChunk` for multi-part or chunked uploads, which allows for resumable uploads and better handling of large files.
- Notifications are used to the let Mux know when the upload is complete (S3 Event Notification) via webhook into a message queue (Kafka, RabbitMQ) to be processed by the workers
- The downloader worker fires off of the new video upload message and downloads the media from S3 to the local storage (SSD or memory if size allows for it)
- The probe worker then analyzes the media and extracts the metadata (bitrate, resolution, frame rate, keyframe intervals, and audio channels). Then performs the calculations for the encoding parameters as well as the chunking strategy.
- The encoding workers then process the media in parallel, generating the different renditions and artifacts. Then upload the result back to S3, and update database accordingly.
