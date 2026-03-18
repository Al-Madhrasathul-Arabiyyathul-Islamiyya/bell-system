namespace ClientApp.Infrastructure.Configuration;

/// <summary>
/// Performs basic validation for client runtime options.
/// </summary>
public sealed class ClientOptionsValidator
{
    readonly StringComparer comparer = StringComparer.Ordinal;

    /// <summary>
    /// Validates the provided options snapshot.
    /// </summary>
    /// <param name="options">The options to validate.</param>
    /// <returns><see langword="true"/> when the options pass basic checks.</returns>
    public bool Validate(ClientOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);

        return options.BackendBaseUrl is not null
            && options.WebSocketUrl is not null
            && comparer.Compare(options.LocalDatabasePath, string.Empty) != 0;
    }
}
