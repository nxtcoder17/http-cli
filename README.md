## http client to speed up GraphQL/rest api development

![demo](https://user-images.githubusercontent.com/22402557/160329835-7f445332-fbbc-48bf-9922-43e907975ba2.gif)

## How to Use (with neovim)

- if you use packer
```lua
  use({
    "nxtcoder17/http-cli",
    run = "pnpm i",  -- npm i, as per your choice
    config = function()
      require("http-cli").setup()
    end,
  })
```

- default config
```lua
local defaultConfig = {
	command = "Gql",
	envFile = function()
		return string.format("%s/%s", vim.env.PWD, "gqlenv.yml")
	end,
}
```

## How to Use

- You should create a gqlenv.yml file in your project root directory, something like this
```yaml
mode: dev
map:
  dev:
    url: <endpoint>
    headers:
      k1: v1
      k2: v2
```

- then, create a file `$filename.yml` file
```yaml
# filename: auth-graphql.yml
---
global:
  email: "sample@gmail.com"

---
query: |
  mutation Login($email: String!, $password: String!) {
    auth {
      login(email: $email, password: $password) {
        id
        userId
        userEmail
      }
    }
  }

variables:
  email: "{{email}}" # this is variable parsing, from either 'gqlenv.json' or from 'global' doc at the top
  password: "hello"  
```

- now, execute it
```sh
go run ./main.go -- $filename $envFileName $lineNumber
```

+ you need to have a variable `url` either in one of the mode vars or global vars

## Inspired By
.http file based REST Client in [Neovim/vim](https://github.com/bayne/vim-dot-http) and Intellij

## Next To Come

- [x] Neovim plugin that could just setup the previous step for you
- [x] i don't know yet 😂
