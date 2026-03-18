using ClientApp.Application.Abstractions;
using ClientApp.Application.Models;

namespace ClientApp.Infrastructure.Services;

/// <summary>
/// Placeholder authenticated session provider.
/// </summary>
public sealed class AuthSession : IAuthSession
{
    /// <inheritdoc />
    public Task<UserSessionModel?> GetCurrentAsync(CancellationToken cancellationToken = default)
    {
        _ = cancellationToken;
        return Task.FromResult<UserSessionModel?>(null);
    }
}
