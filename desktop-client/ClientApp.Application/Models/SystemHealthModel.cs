namespace ClientApp.Application.Models;

/// <summary>
/// Captures the current health of critical client subsystems.
/// </summary>
/// <param name="BackendConnected">Whether the backend HTTP path is available.</param>
/// <param name="RealtimeConnected">Whether realtime updates are connected.</param>
/// <param name="AudioReady">Whether the audio subsystem is ready.</param>
public sealed record SystemHealthModel(bool BackendConnected, bool RealtimeConnected, bool AudioReady);
