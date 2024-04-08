package config

func merge(dst, src *Config) {
	mergeTLS(dst, src)
	mergeStorage(dst, src)
	mergeCommon(dst, src)
}

func mergeCommon(dst, src *Config) {
	mergeStringDef(&dst.ListenAddr, &src.ListenAddr, defaults["listen_addr"])
	mergeStringDef(&dst.GRPCListenAddr, &src.GRPCListenAddr, defaults["grpc_listen_addr"])
	mergeStringDef(&dst.BaseURL, &src.BaseURL, defaults["base_url"])
	mergeString(&dst.RedirectScheme, &src.RedirectScheme)
	mergeStringDef(&dst.JWTSecret, &src.JWTSecret, defaults["jwt_secret"])
	mergeString(&dst.PprofServerAddr, &src.PprofServerAddr)
	mergeBool(&dst.TrustRequestID, &src.TrustRequestID)
	mergeBool(&dst.GRPCReflection, &src.GRPCReflection)
}

func mergeStorage(dst, src *Config) {
	if dst.Storage == nil {
		dst.Storage = src.Storage
	} else if src.Storage != nil {
		mergeStringDef(&dst.Storage.FileStoragePath, &src.Storage.FileStoragePath, defaults["file_storage_path"])
		mergeString(&dst.Storage.DatabaseDSN, &src.Storage.DatabaseDSN)
		mergeBool(&dst.Storage.TraceDB, &src.Storage.TraceDB)
	}
}

func mergeTLS(dst, src *Config) {
	if dst.TLS == nil {
		dst.TLS = src.TLS
	} else if src.TLS != nil {
		mergeBool(&dst.TLS.Enable, &src.TLS.Enable)
		mergeBool(&dst.TLS.UseSelfSigned, &src.TLS.UseSelfSigned)
		mergeString(&dst.TLS.KeyPath, &src.TLS.KeyPath)
		mergeString(&dst.TLS.CertPath, &src.TLS.CertPath)
	}
}

func mergeStringDef(dst, src *string, def string) {
	if *src != "" {
		if *dst == "" || *src != def {
			*dst = *src
		}
	}
}

func mergeString(dst, src *string) {
	if *src != "" {
		*dst = *src
	}
}

func mergeBool(dst, src *bool) {
	if *src {
		*dst = true
	}
}
