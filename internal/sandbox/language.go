package sandbox

// 预定义语言配置
var (
	ProfileGo = Profile{
		Name:       "go",
		Ext:        ".go",
		Image:      "golang:1.21-alpine",
		RunCmd:     []string{"sh", "-c", "go run main.go"},
	}

	ProfilePython = Profile{
		Name:       "python",
		Ext:        ".py",
		Image:      "python:3.12-alpine",
		RunCmd:     []string{"sh", "-c", "python3 main.py"},
	}

	ProfileCpp = Profile{
		Name:       "cpp",
		Ext:        ".cpp",
		Image:      "gcc:13-bookworm",
		CompileCmd: []string{"sh", "-c", "g++ -O2 -std=c++17 -o /code/main main.cpp"},
		RunCmd:     []string{"sh", "-c", "/code/main"},
	}
)

// AllProfiles 所有支持的语言列表
var AllProfiles = []Profile{ProfileGo, ProfilePython, ProfileCpp}

// GetProfile 按名称查找语言配置
func GetProfile(name string) (Profile, bool) {
	for _, p := range AllProfiles {
		if p.Name == name {
			return p, true
		}
	}
	return Profile{}, false
}

// ProfileNames 返回所有语言名称
func ProfileNames() []string {
	names := make([]string, len(AllProfiles))
	for i, p := range AllProfiles {
		names[i] = p.Name
	}
	return names
}
