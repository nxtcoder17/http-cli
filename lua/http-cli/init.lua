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

	-- local vimBufOptions = "vne | setlocal buftype=nofile | setlocal bufhidden=hide | setlocal noswapfile | set ft=json "

	-- local cliExec = string.format("cd %s && go run ./main.go ", pRootDir)
	local cliExec = string.format("%s/bin/http", pRootDir)

	function execute(subCommand)
		local win = vim.api.nvim_get_current_win()
		local buf = vim.api.nvim_create_buf(true, true)
		vim.api.nvim_buf_set_var(buf, "has_stderr", false)

		vim.fn.jobstart({
			cliExec,
			subCommand,
			"--file",
			vim.fn.expand("%:p"),
			"--envFile",
			cfg.envFile(),
			"--lineNo",
			vim.fn.line("."),
		}, {
			stdout_buffered = true,
			on_stdout = function(_, data)
				if data then
					vim.api.nvim_buf_set_lines(buf, -2, -1, false, data)
				end
			end,
			on_stderr = function(_, data)
				if data then
					if #data == 1 and data[1] == "" then
						return
					end
					if not vim.api.nvim_buf_get_var(buf, "has_stderr") then
						vim.api.nvim_buf_set_var(buf, "has_stderr", true)
						vim.api.nvim_buf_set_lines(buf, -1, -1, false, { "", "", "STDERR:", "" })
					end
					vim.api.nvim_buf_set_lines(buf, -1, -1, false, data)
				end
			end,
			on_exit = function(status)
				vim.api.nvim_set_current_win(win)
				vim.cmd("80vne | setlocal buftype=nofile | setlocal bufhidden=hide | setlocal noswapfile")
				vim.cmd("buffer " .. buf)
				vim.cmd("set ft=markdown")
				vim.cmd("norm! G")
			end,
		})
	end

	vim.api.nvim_create_user_command(cfg.graphqlCommand, function()
		execute("graphql")
	end, { desc = "runs http-cli for graphql" })
	vim.api.nvim_create_user_command(cfg.httpCommand, function()
		execute("rest")
	end, { desc = "runs http-cli for rest-api" })
end

return M
