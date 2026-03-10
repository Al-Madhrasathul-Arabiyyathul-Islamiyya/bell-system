namespace ClientApp.Application.Abstractions;

/// <summary>
/// Represents the current authenticated user session.
/// </summary>
public interface IAuthSession
{
    /// <summary>
    /// Gets the active session, if one exists.
    /// </summary>
    /// <param name="cancellationToken">Cancels the lookup.</param>
    /// <returns>The current session or <see langword="null"/>.</returns>
    Task<Models.UserSessionModel?> GetCurrentAsync(CancellationToken cancellationToken = default);
}
