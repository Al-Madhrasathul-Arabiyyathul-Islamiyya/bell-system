using ClientApp.Application.Models;
using System.Collections.ObjectModel;

namespace ClientApp.Application.State;

/// <summary>
/// Holds schedule items loaded into the client runtime.
/// </summary>
public sealed class ScheduleState
{
    /// <summary>
    /// Gets the current schedule items.
    /// </summary>
    public Collection<ScheduleItemModel> Items { get; } = [];
}
