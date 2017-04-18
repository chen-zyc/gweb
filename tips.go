package gweb

import "fmt"

const Version = "V1.0.0"

const logoTpl = `
                    _
                   | |         gweb %s
  __ ___      _____| |__
 / _`+"`"+` \ \ /\ / / _ \ '_ \
| (_| |\ V  V /  __/ |_) |
 \__, | \_/\_/ \___|_.__/
  __/ |
 |___/
`

var Logo = fmt.Sprintf(logoTpl, Version)

const serverInfoTpl = "Server '%s' is now ready to accept connections on port %s.\n"
