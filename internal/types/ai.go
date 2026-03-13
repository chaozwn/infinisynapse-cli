package types

type WebviewMessage struct {
	Type                      string              `json:"type"`
	Text                      string              `json:"text,omitempty"`
	AskResponse               string              `json:"askResponse,omitempty"`
	APIConfiguration          *APIConfiguration    `json:"apiConfiguration,omitempty"`
	Images                    []string            `json:"images,omitempty"`
	AutoApprovalSettings      *AutoApprovalSettings `json:"autoApprovalSettings,omitempty"`
	ChatSettings              *ChatSettings        `json:"chatSettings,omitempty"`
	ChatContent               *ChatContent         `json:"chatContent,omitempty"`
	TaskID                    string              `json:"taskId,omitempty"`
	ConnID                    string              `json:"connId,omitempty"`
	CustomInstructionsSetting string              `json:"customInstructionsSetting,omitempty"`
	SnapshotTs                int64               `json:"snapshotTs,omitempty"`
}

type APIConfiguration struct {
	APIProvider         string            `json:"apiProvider,omitempty"`
	APIModelID          string            `json:"apiModelId,omitempty"`
	APIKey              string            `json:"apiKey,omitempty"`
	OpenAIBaseURL       string            `json:"openAiBaseUrl,omitempty"`
	OpenAIAPIKey        string            `json:"openAiApiKey,omitempty"`
	OpenAIModelID       string            `json:"openAiModelId,omitempty"`
	OpenAINativeAPIKey  string            `json:"openAiNativeApiKey,omitempty"`
	OpenAIHeaders       map[string]string `json:"openAiHeaders,omitempty"`
	InfinisynapseAPIKey string            `json:"infinisynapseApiKey,omitempty"`
	InfinisynapseModel  string            `json:"infinisynapseModelId,omitempty"`
	DeepSeekAPIKey      string            `json:"deepSeekApiKey,omitempty"`
	QwenAPIKey          string            `json:"qwenApiKey,omitempty"`
	QwenAPILine         string            `json:"qwenApiLine,omitempty"`
	ReasoningEffort     string            `json:"reasoningEffort,omitempty"`
}

type AutoApprovalSettings struct {
	MaxRequests            int  `json:"maxRequests,omitempty"`
	MaxSubAgentRequests    int  `json:"maxSubAgentRequests,omitempty"`
	DatabaseReturnLimit    int  `json:"databaseReturnLimit,omitempty"`
	DelegateMaxConcurrency int  `json:"delegateMaxConcurrency,omitempty"`
	EnableNotifications    bool `json:"enableNotifications,omitempty"`
	DebugMode              bool `json:"debugMode,omitempty"`
	EnableWebSearch        bool `json:"enableWebSearch,omitempty"`
	EnableReadImage        bool `json:"enableReadImage,omitempty"`
	EnableBrowser          bool `json:"enableBrowser,omitempty"`
	EnableMap              bool `json:"enableMap,omitempty"`
}

type ChatSettings struct {
	Mode string `json:"mode,omitempty"`
}

type ChatContent struct {
	Message string   `json:"message,omitempty"`
	Images  []string `json:"images,omitempty"`
}

type ExtensionState struct {
	APIConfiguration     *APIConfiguration      `json:"apiConfiguration,omitempty"`
	CustomInstructions   string                 `json:"customInstructions,omitempty"`
	CurrentTaskItem      map[string]interface{} `json:"currentTaskItem,omitempty"`
	InfiniMessages       []interface{}          `json:"infiniMessages,omitempty"`
	AutoApprovalSettings *AutoApprovalSettings  `json:"autoApprovalSettings,omitempty"`
	ChatSettings         *ChatSettings          `json:"chatSettings,omitempty"`
}

type AISettingsUpdate struct {
	APIConfiguration          *APIConfiguration      `json:"apiConfiguration,omitempty"`
	CustomInstructionsSetting string                 `json:"customInstructionsSetting,omitempty"`
	AutoApprovalSettings      *AutoApprovalSettings  `json:"autoApprovalSettings,omitempty"`
	TaskID                    string                 `json:"taskId,omitempty"`
}

type MessageResponse struct {
	Success       bool           `json:"success,omitempty"`
	State         *ExtensionState `json:"state,omitempty"`
	Notification  string         `json:"notification,omitempty"`
	OpenAIModels  []interface{}  `json:"openAiModels,omitempty"`
	ForceReplace  bool           `json:"forceReplace,omitempty"`
}
