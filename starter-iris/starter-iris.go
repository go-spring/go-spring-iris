/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package StarterIris

import (
	"context"
	"fmt"

	"github.com/go-spring/spring-boot"
	"github.com/go-spring/spring-iris/spring-iris"
	"github.com/go-spring/spring-logger"
	"github.com/go-spring/spring-utils"
	"github.com/go-spring/spring-web"
	"github.com/kataras/iris/v12"
)

func init() {
	SpringBoot.RegisterNameBeanFn("iris-app-starter", func(config IrisServerConfig) *IrisAppStarter {
		return &IrisAppStarter{app: iris.New(), cfg: config}
	})
}

// IrisServerConfig Iris 服务器配置
type IrisServerConfig struct {
	EnableHTTP  bool   `value:"${iris.server.enable:=true}"`      // 是否启用 HTTP
	Port        int    `value:"${iris.server.port:=8080}"`        // HTTP 端口
	EnableHTTPS bool   `value:"${iris.server.ssl.enable:=false}"` // 是否启用 HTTPS
	SSLPort     int    `value:"${iris.server.ssl.port:=8443}"`    // SSL 端口
	SSLCert     string `value:"${iris.server.ssl.cert:=}"`        // SSL 证书
	SSLKey      string `value:"${iris.server.ssl.key:=}"`         // SSL 秘钥
}

// IrisAppStarter
type IrisAppStarter struct {
	_ SpringBoot.ApplicationEvent `export:""`

	app *iris.Application
	cfg IrisServerConfig
}

func (starter *IrisAppStarter) OnStartApplication(ctx SpringBoot.ApplicationContext) {

	for _, mapping := range SpringIris.DefaultWebMapping.Mappings {
		if mapping.Matches(ctx) {
			filters := mapping.Filters()
			for _, s := range mapping.FilterNames() {
				var f SpringWeb.Filter
				ctx.GetBean(&f, s)
				filters = append(filters, f)
			}
			for _, method := range SpringWeb.GetMethod(mapping.Method()) {
				starter.app.Handle(method, mapping.Path(), mapping.Handler())
			}
			//c.Request(mapping.Method(), mapping.Path(), mapping.Handler(), filters...)
		}
	}

	go func() {
		var r iris.Runner
		address := fmt.Sprintf(":%d", starter.cfg.Port)
		if starter.cfg.EnableHTTPS {
			r = iris.TLS(address, starter.cfg.SSLCert, starter.cfg.SSLKey)
		} else {
			r = iris.Addr(address)
		}
		err := starter.app.Run(r, iris.WithoutServerError(iris.ErrServerClosed))
		SpringLogger.Infof("exit http server on %s return %v", address, err)
	}()
}

func (starter *IrisAppStarter) OnStopApplication(ctx SpringBoot.ApplicationContext) {
	err := starter.app.Shutdown(context.TODO())
	SpringUtils.Panic(err).When(err != nil)
}
