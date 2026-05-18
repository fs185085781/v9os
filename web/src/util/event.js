import mitt from "mitt";
import { onUnmounted } from "vue";
const emitter = mitt();
window.addEventListener("message", (event) => {
  const msg = event.data;
  if (msg.action) {
    emitter.emit(msg.action, msg.data);
  }
});
export function useEventBus(eventName, callback) {
  emitter.on(eventName, callback);
  onUnmounted(() => {
    emitter.off(eventName, callback);
  });
}
export default emitter;
