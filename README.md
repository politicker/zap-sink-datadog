## Datadog Logs HTTP API sink for Zap

This makes a blocking network call every time you log. Probably not a good idea for any production environment (you 
should use the agent in prod anyway.)

Most useful for situations where, for whatever reason, you don't want to set up the datadog agent yet.
