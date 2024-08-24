stream: 
	rm -rf ./stream_hls
	mkdir -p ./stream_hls
	yt-dlp -o - "https://www.youtube.com/watch?v=d9T5EsiuVhc" | ffmpeg -i pipe:0 -c:v libx264 -preset veryfast -g 50 -hls_time 10 -hls_list_size 0 \
-hls_segment_filename "./stream_hls/segment_%03d.ts" \
-f hls ./stream_hls/stream.m3u8