local M = {}

local defaultConfig = {
  httpCommand = "Http",
  graphqlCommand = "Gql",
  envFile = function()
    return string.format("%s/%s", vim.env.PWD, "gqlenv.yml")
  end,
}

M.setup = function(cfg)
  cfg = vim.tbl_deep_extend("force", defaultConfig, cfg or {})

  local currDir = debug.getinfo(1).source:match("@?(.*/)")
  local pRootDir = currDir .. "../.."

  local vimBufOptions = "vne | setlocal buftype=nofile | setlocal bufhidden=hide | setlocal noswapfile | set ft=json "

  --local cliExec = string.format("cd %s && go run ./main.go ", pRootDir)
  local cliExec = string.format("%s/bin/http", pRootDir)

  local gqlCmd = string.format(
    "command! -nargs=0 %s execute '%s | r! %s graphql --file ' . expand('%%:p') . ' -envFile %s --lineNo ' . line('.')",
    cfg.graphqlCommand,
    vimBufOptions,
    cliExec,
    cfg.envFile()
  )
  vim.cmd(gqlCmd)

  local httpCmd = string.format(
    "command! -nargs=0 %s execute '%s | r! %s rest --file ' . expand('%%:p') . ' -envFile %s --lineNo ' . line('.')",
    cfg.httpCommand,
    vimBufOptions,
    cliExec,
    cfg.envFile()
  )
  vim.cmd(httpCmd)
end

return M
