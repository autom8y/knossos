package slack

// Progressive response streaming for Slack.
//
// Slack's streaming API (chat.startStream/appendStream/stopStream) provides
// progressive text rendering in threads. As of the current slack-go version,
// these APIs are not natively supported.
//
// Current behavior: The handler posts a complete message via PostMessage after
// pipeline processing completes. This is functionally correct but lacks the
// progressive UX of streaming.
//
// TODO: When slack-go adds native streaming support or when the Slack streaming
// API stabilizes, implement progressive rendering:
//
//  1. On pipeline start: call chat.startStream to create a streaming message
//  2. As chunks arrive: call appendStream with text increments (respecting
//     StreamChunkSize and StreamChunkDelay from SlackConfig)
//  3. On pipeline complete: call stopStream to finalize the message
//  4. On error: call stopStream with error indicator
//
// The SlackConfig.StreamingEnabled, StreamChunkSize, and StreamChunkDelay
// fields are ready for this implementation.
