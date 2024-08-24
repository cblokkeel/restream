stream:
	@if [ -z "$(URL)" ]; then \
		echo "Please provide a YouTube URL using URL=<youtube_url>"; \
		exit 1; \
	fi
	rm -rf ./stream_hls
	mkdir -p ./stream_hls
	(yt-dlp -o - "$(URL)" | ffmpeg -i pipe:0 -c:v libx264 -preset veryfast -g 50 -hls_time 10 -hls_list_size 0 \
	-hls_segment_filename "./stream_hls/segment_%03d.ts" \
	-f hls ./stream_hls/stream.m3u8) &
	go run ./subtitles/main.go &
	sleep 10
	cd vite-project && npm run dev