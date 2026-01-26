```bash
åååååxxxxxxxxxx # 完整示例：批量分析并生成报告export AGENTBOX_API_KEY=$(email-analyzer login 2>/dev/null | grep Token | awk '{print $2}')email-analyzer analyze \  -d /path/to/emails/ \  -w 10 \  -t 15m \  -f json \  -o analysis_report.jsonecho "分析完成，报告保存到 analysis_report.json"bash
```
