const WEpatcher__electron = require("electron");
const WEpatcher__path = require("path");
const WEpatcher__instanceArg = process.argv.find(a => a.startsWith("--instance="));
const WEpatcher__instanceName = WEpatcher__instanceArg ? WEpatcher__instanceArg.split("=")[1] : "modified";
WEpatcher__electron.app.setPath('userData', WEpatcher__path.join(
  WEpatcher__electron.app.getPath('appData'),
  WEpatcher__electron.app.getName() + "-" + WEpatcher__instanceName
));
