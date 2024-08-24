<template>
  <div>
    <video ref="videoPlayer" controls style="width: 100%;"></video>
  </div>
</template>

<script>
import Hls from "hls.js";

export default {
  name: "VideoPlayer",
  mounted() {
    this.setupHls();
  },
  methods: {
    setupHls() {
      const video = this.$refs.videoPlayer;

      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource("http://localhost:1172/stream.m3u8");
        hls.attachMedia(video);
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          video.play();
        });
      } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
        video.src = "http://localhost:1171/stream.m3u8";
        video.addEventListener("loadedmetadata", () => {
          video.play();
        });
      }
    },
  },
};
</script>

<style scoped>
/* Add any necessary styles here */
</style>
