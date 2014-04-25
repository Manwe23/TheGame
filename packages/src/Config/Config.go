package Config

const DatabaseHost = "localhost"
const DatabasePort = 3306
const DatabaseUser = "engine"
const DatabasePassword = "enginepassword"

const MapDatabaseName = "mapconfigurationdata"

const EngineClientInboxSize = 1000
const ClientEngineInboxSize = 1000

const ClientProcessorRequestsQueueSize = 1000
const ClientProcessorInboxSize = 1000

const ModuleInboxSize = 1000
const ModuleControlChanSize = 10

const HttpServerPort = 80
const HttpServerDir = "web/"

const EngineResponsesQueueSize = 1000
const EngineRequestsQueueSize = 1000
const EngineInboxSize = 1000
const EngineSendMessageTimeout = 1 // seconds
