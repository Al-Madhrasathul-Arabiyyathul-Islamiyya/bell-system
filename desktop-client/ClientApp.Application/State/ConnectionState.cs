namespace ClientApp.Application.State;

/// <summary>
/// Tracks the current backend and realtime connectivity state.
/// </summary>
public sealed class ConnectionState
{
    /// <summary>
    /// Gets or sets whether the backend HTTP API is reachable.
    /// </summary>
    public bool IsBackendConnected { get; set; }

    /// <summary>
    /// Gets or sets whether the realtime channel is connected.
    /// </summary>
    public bool IsRealtimeConnected { get; set; }
}
