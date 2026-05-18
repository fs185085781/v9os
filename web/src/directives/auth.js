import { useStore } from "@/stores/user";
let store = null;
const getStore = () => {
  if (!store) {
    store = useStore();
  }
  return store;
};
const checkPermission = (el, binding) => {
  if (!el) {
    return;
  }
  const parentNode = el.parentNode;
  if (!parentNode) {
    return;
  }
  const requiredPermissions = binding.value;
  const auth = getStore().auths;
  let hasPermission = auth.auths.includes("all");
  if (!hasPermission) {
    if (typeof requiredPermissions === "string") {
      hasPermission = auth.auths.includes(requiredPermissions);
    } else if (
      Array.isArray(requiredPermissions) &&
      requiredPermissions.length > 0
    ) {
      hasPermission = requiredPermissions.every((perm) =>
        auth.auths.includes(perm),
      );
    } else {
      hasPermission = false;
    }
  }
  if (!hasPermission) {
    if (auth.init) {
      if (el.parentNode) {
        el.parentNode.removeChild(el);
      }
    } else {
      el.style.display = "none";
    }
  } else {
    el.style.display = "";
  }
};
export default {
  mounted(el, binding) {
    checkPermission(el, binding);
  },
  updated(el, binding) {
    checkPermission(el, binding);
  },
};
export function checkAuth(requiredPermission) {
  const auth = getStore().auths;
  let hasPermission = auth.auths.includes("all");
  return hasPermission || auth.auths.includes(requiredPermission);
}
