namespace ClientApp.Infrastructure.Logging;

/// <summary>
/// Placeholder logging adapter for application events.
/// </summary>
public sealed class AppLoggerAdapter
{
    /// <summary>
    /// Gets the time of the most recent log write.
    /// </summary>
    public DateTimeOffset? LastLoggedAt { get; private set; }

    /// <summary>
    /// Writes an informational message.
    /// </summary>
    /// <param name="message">The message to record.</param>
    public void LogInformation(string message)
    {
        _ = message;
        LastLoggedAt = DateTimeOffset.UtcNow;
    }
}
