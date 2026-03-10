namespace ClientApp.Infrastructure.Http;

/// <summary>
/// Placeholder factory for authenticated backend HTTP clients.
/// </summary>
public sealed class AuthenticatedHttpClientFactory
{
    /// <summary>
    /// Gets the default HTTP version used by created clients.
    /// </summary>
    public Version DefaultVersion { get; } = new(2, 0);

    /// <summary>
    /// Creates a new HTTP client instance.
    /// </summary>
    public HttpClient CreateClient() =>
        new()
        {
            DefaultRequestVersion = DefaultVersion,
        };
}
