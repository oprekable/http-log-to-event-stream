local http = require "resty.http"
local json = require('cjson')

local TokenHandler = {
    VERSION = "1.0",
    PRIORITY = 1000,
}

local function fetch_token(conf, access_token)
    local http_client = http:new()
    local res, err = http_client:request_uri(conf.my_auth_endpoint, {
        method = "GET",
        ssl_verify = false,
        headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer " .. access_token }
    })

    if not res then
        kong.log.err("failed to call new fetch endpoint: ", err)
        return kong.response.exit(500)
    end

    if res.status == 401 then
        kong.log.err("new fetch endpoint responded with status: ", res.status)
        return kong.response.exit(res.status, res.body, {
            ["Content-Type"] = "application/json",
        })
    end

    if res.status ~= 200 then
        kong.log.err("new fetch endpoint responded with status: ", res.status)
        return kong.response.exit(500)
    end

    return json.decode(res.body)
end

local function manage_access_token(conf, access_token)
    local auth_body_introspection
    local user_id, email

    auth_body_introspection = fetch_token(conf, access_token)
    user_id = auth_body_introspection.data.user_id
    email = auth_body_introspection.data.email

    kong.service.request.add_header("X-User-Id", user_id)
    kong.service.request.add_header("X-User-Email", email)
end

function TokenHandler:access(conf)
    local access_token = ngx.req.get_headers()[conf.token_header]

    if not access_token then
        kong.response.exit(401)
    end

    access_token = access_token:sub(8, -1)
    manage_access_token(conf, access_token)
end

return TokenHandler
