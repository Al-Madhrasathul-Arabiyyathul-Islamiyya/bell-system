namespace ClientApp.Application.Abstractions;

/// <summary>
/// Reads schedule data from the backend HTTP API.
/// </summary>
public interface IScheduleApi
{
    /// <summary>
    /// Fetches a lightweight schedule summary for the client dashboard.
    /// </summary>
    /// <param name="cancellationToken">Cancels the HTTP request.</param>
    /// <returns>A summary string representing the current schedule state.</returns>
    Task<string> GetScheduleSummaryAsync(CancellationToken cancellationToken = default);
}
