using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Services;

/// <summary>
/// Placeholder WebSocket-based realtime update implementation.
/// </summary>
public sealed class WebSocketRealtimeUpdates : IRealtimeUpdates
{
    /// <inheritdoc />
    public Task StartAsync(CancellationToken cancellationToken = default) =>
        Task.CompletedTask;
}
