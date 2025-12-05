local json = require("json")

-- delete Host after some time (add a few seconds to remove any race-conditioning)
local host_server_alive_time = 65

-- interval for stale cleanup
local host_server_stale_cleanup_interval = 5

-- mapping table: host.domain.com -> [ { serverName: "server1", aliveUntil: 600 }, { serverName: "server2", aliveUntil: 300 }, ... ]
local host_servers = host_servers or {}

-- helper to update host-server mapping
local function apply_diff(serverName, diff)
    -- Add hosts
    if diff.added then
        local currentTime = core.now()

        for _, host in ipairs(diff.added) do
            host_servers[host] = host_servers[host] or {}

            local exists = false
            for _, server in ipairs(host_servers[host]) do
                if server.serverName == serverName then exists = true break end
            end

            local serverObject = {
                serverName = serverName,
                aliveUntil = currentTime.sec + host_server_alive_time
            }

            if not exists then
                core.Info("Added Server "..serverName.." in "..host..". Reason: Diff-Added")

                table.insert(host_servers[host], serverObject)
            else
                -- Update Alive Time
                for i = 1, #host_servers[host] do
                    local server = host_servers[host][i]

                    if server.serverName == serverName then
                        core.Info("Bumped "..serverName.."'s Alive Time in "..host.." by "..host_server_alive_time..". Reason: Diff-Added + Update")

                        host_servers[host][i] = serverObject
                    end
                end
            end
        end
    end

    -- Remove hosts
    if diff.removed then
        for _, host in ipairs(diff.removed) do
            if host_servers[host] then
                for i = #host_servers[host], 1, -1 do
                    local server = host_servers[host][i]

                    if server.serverName == serverName then
                        core.Info("Removed Server "..serverName.." from "..host..". Reason: Diff-Removed")

                        table.remove(host_servers[host], i)
                    end
                end
                if #host_servers[host] == 0 then
                    host_servers[host] = nil
                end
            end
        end
    end
end

-- helper to send HTTP Text Responses
local function sendTextResponse(txn, msg, status)
    txn:set_status(status)
    txn:add_header("Content-Type", "text/plain")
    txn:start_response()
    txn:send(msg)
    return
end

-- Stale Handler
local function removeStale()
    local currentTime = core.now()
    
    for host, servers in pairs(host_servers) do
        for i = #servers, 1, -1 do
            local server = servers[i]

            if server.aliveUntil < currentTime.sec then
                core.Info("Removed Server "..server.serverName.." from "..host..". Reason: Stale")

                table.remove(host_servers[host], i)
            end
        end
        if #host_servers[host] == 0 then
            host_servers[host] = nil
        end
    end
end

function cleanup_stales()
    while true do
        core.msleep(host_server_stale_cleanup_interval * 1000)

        removeStale()
    end
end

-- register Stale-Cleanup task
core.register_task(cleanup_stales)

-- HTTP endpoint handler
function handle_update(txn)
    local body = txn:receive()
    if not body or #body == 0 then
        sendTextResponse(txn, "Empty body request", 400)
        return
    end

    local payload, err = json.decode(body)
    if not payload then
        sendTextResponse(txn, "Invalid JSON "..(err or "unknown"), 400)
        return
    end

    if not payload.serverName or not payload.diff then
        sendTextResponse(txn, "Missing serverName or diff", 400)
        return
    end

    apply_diff(payload.serverName, payload.diff)

    sendTextResponse(txn, "Mapping successfully updated!", 200)

    core.Info("Mapping successfully updated!")
end

-- register the HTTP endpoint
core.register_service("update_mapping", "http", handle_update)

-- resolve backend for a given host
function resolve_backend(txn)
    local host = txn:get_var("txn.sni")

    if host == nil or host == "" then
        txn:set_var("txn.backend", "default_backend")

        core.Alert("No SNI found, using default backend")

        return
    end

    local servers = host_servers[host]

    if not servers or #servers == 0 then
        txn:set_var("txn.backend", "default_backend")

        core.Info("No Server found, using default backend")
    elseif #servers == 1 then
        txn:set_var("txn.backend", servers[1].serverName.."_backend")

        core.Info("Selected "..servers[1].serverName.." as Backend")
    elseif #servers >= 1 then
        local server_names = {}
        for i, server in ipairs(servers) do
            server_names[i] = server.serverName
        end
        
        -- multiple servers -> failover backend
        local backend_name = table.concat(server_names, "_") .. "_failover_backend"
        txn:set_var("txn.backend", backend_name)

        core.Info("Selected "..backend_name.." as Failover Backend")
    end
end

core.register_action("resolve_backend", {"tcp-req", "http-req" }, resolve_backend)
