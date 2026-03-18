namespace ClientApp.Infrastructure.Configuration;

/// <summary>
/// Defines the runtime configuration required by the desktop client.
/// </summary>
public sealed class ClientOptions
{
    /// <summary>
    /// Gets or sets the backend base URL.
    /// </summary>
    public Uri? BackendBaseUrl { get; set; }

    /// <summary>
    /// Gets or sets the WebSocket URL.
    /// </summary>
    public Uri? WebSocketUrl { get; set; }

    /// <summary>
    /// Gets or sets the local SQLite database path.
    /// </summary>
    public string LocalDatabasePath { get; set; } = string.Empty;
}
