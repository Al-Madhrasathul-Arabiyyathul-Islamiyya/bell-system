using ClientApp.Application.Abstractions;

namespace ClientApp.Application.UseCases;

/// <summary>
/// Refreshes schedule data from the backend API.
/// </summary>
public sealed class RefreshScheduleUseCase(IScheduleApi scheduleApi)
{
    /// <summary>
    /// Loads the latest schedule summary.
    /// </summary>
    /// <param name="cancellationToken">Cancels the operation.</param>
    public Task<string> ExecuteAsync(CancellationToken cancellationToken = default) =>
        scheduleApi.GetScheduleSummaryAsync(cancellationToken);
}
