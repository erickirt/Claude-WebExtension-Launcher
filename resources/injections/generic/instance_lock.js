const __WEpatcher_electron = require("electron");
function __modifiedLock() {
  const __instanceArg = process.argv.find(a => a.startsWith("--instance="));
  const __instanceName = __instanceArg ? __instanceArg.split("=")[1] : "modified";
  __WEpatcher_electron.app.setName(__WEpatcher_electron.app.getName() + "-" + __instanceName);
  const got = __WEpatcher_electron.app.requestSingleInstanceLock();
  __WEpatcher_electron.app.setName(__WEpatcher_electron.app.getName().replace("-" + __instanceName, ""));
  return got;
}
