import { createApp, createSSRApp } from "vue";
import App from "./App.vue";
import "./style.css";

const prerendered = document.getElementById("app")?.hasChildNodes();
(prerendered ? createSSRApp : createApp)(App).mount("#app");
