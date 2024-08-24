<template>
  <div>
    <video ref="videoPlayer" controls style="width: 100%;">
      <track kind="subtitles" :src="subtitlesUrl" srclang="fr" label="French" default>
    </video>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Hls from "hls.js";

const videoPlayer = ref(null);
const subtitlesUrl = ref("/api/subtitles.vtt");

const setupHls = () => {
  const video = videoPlayer.value;

  if (Hls.isSupported()) {
    const hls = new Hls();
    hls.loadSource("/api/stream.m3u8");
    hls.attachMedia(video);
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      video.play();
    });
  } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
    video.src = "/api/stream.m3u8";
    video.addEventListener("loadedmetadata", () => {
      video.play();
    });
  }
};

const setupSubtitles = () => {
  const video = videoPlayer.value;
  
  setInterval(() => {
    subtitlesUrl.value = `/api/subtitles.vtt?t=${Date.now()}`;
    
    setTimeout(() => {
      if (video.textTracks[0]) {
        video.textTracks[0].mode = 'showing';
      }
    }, 100);
  }, 10000); // Reload every 10 seconds
};

onMounted(() => {
  setupHls();
  setupSubtitles();
});
</script>