namespace ClientApp.Infrastructure.Persistence.Entities;

/// <summary>
/// Represents a cached schedule snapshot stored locally.
/// </summary>
public sealed class CachedScheduleEntity
{
    /// <summary>
    /// Gets or sets the cache entry identifier.
    /// </summary>
    public Guid Id { get; set; }

    /// <summary>
    /// Gets or sets the serialized payload.
    /// </summary>
    public string Payload { get; set; } = string.Empty;
}
