namespace ClientApp.Infrastructure.Realtime;

/// <summary>
/// Placeholder router for incoming realtime messages.
/// </summary>
public sealed class RealtimeMessageRouter
{
    /// <summary>
    /// Gets the last routed message payload.
    /// </summary>
    public string? LastMessage { get; private set; }

    /// <summary>
    /// Routes a raw message payload.
    /// </summary>
    /// <param name="message">The incoming message payload.</param>
    public void Route(string message) => LastMessage = message;
}
