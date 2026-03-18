namespace ClientApp.Application.Abstractions;

/// <summary>
/// Provides the current time for schedule and countdown calculations.
/// </summary>
public interface IClock
{
    /// <summary>
    /// Gets the current local time.
    /// </summary>
    DateTimeOffset Now { get; }
}
