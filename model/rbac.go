package model

type UserRole string

const (
    RoleAdmin  UserRole = "admin"
    RoleUser   UserRole = "user"
    RoleViewer UserRole = "viewer"
)

type Permission struct {
    Dashboard   bool `json:"dashboard"`
    FileManager bool `json:"file_manager"`
    AppStore    bool `json:"app_store"`
    WebTerminal bool `json:"web_terminal"`
    Settings    bool `json:"settings"`
}

func GetPermissionsForRole(role UserRole) Permission {
    switch role {
    case RoleAdmin:
        return Permission{Dashboard: true, FileManager: true, AppStore: true, WebTerminal: true, Settings: true}
    case RoleUser:
        return Permission{Dashboard: true, FileManager: true, AppStore: false, WebTerminal: false, Settings: false}
    case RoleViewer:
        return Permission{Dashboard: true, FileManager: false, AppStore: false, WebTerminal: false, Settings: false}
    default:
        return Permission{Dashboard: true}
    }
}

type FirstBootSetup struct {
    Username string `json:"username" validate:"required,min=3,max=32"`
    Password string `json:"password" validate:"required,min=6"`
}
