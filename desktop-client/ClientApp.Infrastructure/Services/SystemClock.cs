using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Services;

/// <summary>
/// Default system clock implementation.
/// </summary>
public sealed class SystemClock : IClock
{
    /// <inheritdoc />
    public DateTimeOffset Now => DateTimeOffset.Now;
}
