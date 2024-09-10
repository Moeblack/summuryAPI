package apis

import (
    "github.com/gin-gonic/gin"
    "github.com/sashabaranov/go-openai"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "net/http"
    "simple-one-api/pkg/mylog"
    "go.uber.org/zap"
)

type SummarizeConfig struct {
    Summarize struct {
        Model              string `yaml:"model"`
        MaxInputLength     int    `yaml:"max_input_length"`
        MaxSummaryLength   int    `yaml:"max_summary_length"`
        PromptTemplateFile string `yaml:"prompt_template_file"`
    } `yaml:"summarize"`
}

var summarizeConfig SummarizeConfig

func init() {
    // 读取配置文件
    configData, err := ioutil.ReadFile("summarize_config.yaml")
    if err != nil {
        mylog.Logger.Error("Failed to read summarize config", zap.Error(err))
        return
    }

    err = yaml.Unmarshal(configData, &summarizeConfig)
    if err != nil {
        mylog.Logger.Error("Failed to parse summarize config", zap.Error(err))
        return
    }
}

func SummarizeHandler(c *gin.Context) {
    // 读取摘要提示模板
    promptTemplate, err := ioutil.ReadFile(summarizeConfig.Summarize.PromptTemplateFile)
    if err != nil {
        mylog.Logger.Error("Failed to read summarize prompt template", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // 解析请求体
    var req struct {
        Text string `json:"text" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // 检查输入长度
    if len(req.Text) > summarizeConfig.Summarize.MaxInputLength {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Input text too long"})
        return
    }

    // 构造 OpenAI 格式的请求
    oaiReq := openai.ChatCompletionRequest{
        Model: summarizeConfig.Summarize.Model,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    "system",
                Content: string(promptTemplate),
            },
            {
                Role:    "user",
                Content: req.Text,
            },
        },
        MaxTokens: summarizeConfig.Summarize.MaxSummaryLength,
    }

    // 将构造好的请求设置到上下文中
    c.Set("chatCompletionRequest", &oaiReq)

    // 修改请求路径以匹配 OpenAIHandler 的处理条件
    c.Request.URL.Path = "/v1/chat/completions"

    // 调用 OpenAIHandler
    OpenAIHandler(c)
}