#######################################################################################################
# Set your parameters
#######################################################################################################

# local
# upstreams
env=local
upstream_proxy="172.16.100.252:8020"
upstream_rainbow_app_service="172.16.100.252:8081"
upstream_logs_service="172.16.100.252:19080"

apisix_addr=http://127.0.0.1:9180
servers_domain=nftrainbow.me
rainbow_api_addr=http://172.16.100.252:8080
settle_addr=http://172.16.100.252:8091
apikey_confura_main="0rW8CEuqNvDaWNybiukVXK5kJp9GP3rdptimpqxu9bdc"
apikey_confura_test="0djrpfkthikrMfSkRzHDdAVD6biYJ42GaWopMkew3t6"
apikey_scan_main="xxx"
apikey_scan_test="xxx"

# dev
# env=dev
# apisix_addr=http://dev-apisix-admin.nftrainbow.cn
# servers_domain=nftrainbow.cn
# rainbow_api_addr=http://127.0.0.1:8080
# settle_addr=http://127.0.0.1:8091
# upstream_proxy="172.18.0.1:8020"
# apikey_confura_main="0rW8CEuqNvDaWNybiukVXK5kJp9GP3rdptimpqxu9bdc"
# apikey_confura_test="0djrpfkthikrMfSkRzHDdAVD6biYJ42GaWopMkew3t6"

echo "开始配置apisix路由"

#######################################################################################################

# ******************** rainbow 使用的upstream *******************

# 添加upstream
curl $apisix_addr/apisix/admin/upstreams/100 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -i -X PUT -d '
{
    "type":"roundrobin",
    "nodes":{
        "'${upstream_proxy}'": 1
    }
}'

curl $apisix_addr/apisix/admin/upstreams/200 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -i -X PUT -d '
{
    "type":"roundrobin",
    "nodes":{
        "'${upstream_rainbow_app_service}'": 1
    }
}'

curl $apisix_addr/apisix/admin/upstreams/300 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -i -X PUT -d '
{
    "type":"roundrobin",
    "nodes":{
        "'${upstream_logs_service}'": 1
    }
}'

# 查upstream
curl $apisix_addr/apisix/admin/upstreams -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1'

# ******************** rainbow 使用的路由 *******************

# rainbow open api
curl $apisix_addr/apisix/admin/routes/1000 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-openapi",
  "desc": "rainbow open api 路由，只匹配openapi需要收费的api",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/v1/(accounts|mints|transfers|burns|contracts|metadata|files|nft|tx)/.*$"]
  ],
  "host": "devapi.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"jwt-auth", "value":"{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-api\",\"env\":\"'${env}'\"}"},
         {"name":"rainbow-api-parser", "value":"{}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"request\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "'${rainbow_api_addr}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"count","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# exit 0

# TODO: rainbow api dashboard 收费相关接口
curl $apisix_addr/apisix/admin/routes/1100 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-dashboard-api",
  "desc": "rainbow dashboard api 路由,只匹配dashboard需要收费的api",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/dashboard/apps/.*/(contracts|nft).*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "methods": ["POST"],
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"jwt-auth", "value":"{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-dashboard\",\"env\":\"'${env}'\"}"},
         {"name":"rainbow-api-parser", "value":"{}"},
         {"name":"count", "value":"{}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "'${rainbow_api_addr}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"count","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# rainbow api dashboard 免费接口
curl $apisix_addr/apisix/admin/routes/1120 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-dashboard-api-free",
  "desc": "rainbow dashboard api 路由,只匹配dashboard免费api",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/dashboard/.*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "methods": ["POST"],
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"jwt-auth", "value":"{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-dashboard\",\"env\":\"local\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "'${rainbow_api_addr}'"
      }
    }
  },
  "upstream_id": "100",
  "priority": 300
}'

# rainbow apps
curl $apisix_addr/apisix/admin/routes/1130 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-app-service",
  "desc": "rainbow-app-service 路由",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/apps/.*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "upstream_id": "200",
  "priority": 400
}'

# rainbow apps
curl $apisix_addr/apisix/admin/routes/1140 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-http-logs",
  "desc": "rainbow-http-logs 路由",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/logs/.*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "upstream_id": "300",
  "priority": 400
}'

# settle 服务
curl $apisix_addr/apisix/admin/routes/1150 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-settle",
  "desc": "rainbow settle",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/settle/.*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "plugins": {
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "'${settle_addr}'"
      }
    }
  },
  "upstream_id": "100",
  "priority": 400
}'

# rainbow api 其它所有接口，包括 v1其它,swagger,debug,dashboard,settle,admin
curl $apisix_addr/apisix/admin/routes/1200 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-api-normal",
  "desc": "rainbow api 基础路由，优先级最低，用于免费接口",
  "uri": "/*",
  "hosts": ["devapi.'${servers_domain}'","dev.'${servers_domain}'","devadmin.'${servers_domain}'"],
  "plugins": {
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "'${rainbow_api_addr}'"
      }
    }
  },
  "upstream_id": "100",
  "priority": 0
}'

# ******************** confura 路由 ********************
# cspace-main
curl $apisix_addr/apisix/admin/routes/2000 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-cspace-main",
  "desc": "confura core space main net",
  "uri": "/*",
  "host": "dev-rpc-cspace-main.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":true,\"is_cspace\":true}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Url": "https://main.confluxrpc.com/'${apikey_confura_main}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# cspace-test
curl $apisix_addr/apisix/admin/routes/2100 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-cspace-test",
  "desc": "confura core space test net",
  "uri": "/*",
  "host": "dev-rpc-cspace-test.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":false,\"is_cspace\":true}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Url": "https://test.confluxrpc.com/'${apikey_confura_test}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# espace-main
curl $apisix_addr/apisix/admin/routes/2200 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-espace-main",
  "desc": "confura espace mainnet",
  "uri": "/*",
  "host": "dev-rpc-espace-main.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":true,\"is_cspace\":false}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Url": "https://evm.confluxrpc.com/'${apikey_confura_main}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# espace-test
curl $apisix_addr/apisix/admin/routes/2300 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-espace-test",
  "desc": "confura espace testnet",
  "uri": "/*",
  "host": "dev-rpc-espace-test.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":false,\"is_cspace\":false}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Url": "https://evmtestnet.confluxrpc.com/'${apikey_confura_test}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# Scan
# cspace-main
curl $apisix_addr/apisix/admin/routes/3000 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "scan-cspace-main",
  "desc": "scan core space main net",
  "uri": "/*",
  "host": "dev-scan-cspace-main.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"header\"}"},
         {"name":"scan-parser", "value":"{\"is_mainnet\":true,\"is_cspace\":true}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "X-Rainbow-Target-Addr": "https://api.confluxscan.net",
        "X-Rainbow-Append-Query": "'apiKey=${apikey_scan_main}'",
        "apiKey": "'${apikey_scan_main}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"count", "value":"{}"}
       ]
    }
  },
  "upstream_id": "100",
  "priority": 400
}'

echo "配置apisix路由完成"
# *************************** 证书相关 ***********************************

# ssh证书生成

# openssl req -new -out server.csr -key server.key -subj "/C=CN/ST=BeiJing/L=BeiJing/O=blockchain/OU=conflux/CN=api.rainbow.com

# # ***************************** DEV 环境 ********************************
# 1. 将 $servers_domain 修改为 nftrainbow.cn
# 2. 将 127.0.0.1:9180 修改为 dev-apisix-admin.nftrainbow.cn
# 3. 将 upstream 修改为 172.18.0.1:8020
# 4. rainbow-api request-rewrite的header 修改为 172.18.0.1.8080
# 5. plugins 增加 http logger
#     "http-logger": {
#       "_meta": {
#         "disable": false
#       },
#       "include_req_body": true,
#       "include_resp_body": true,
#       "uri": "http://172.18.0.1:19080/logs/rconsole"
#     }
