import { Link } from 'react-router-dom';

import { terminalMonitoringHref } from './monitoring-routes.js';
import { streamConnectionStatusClass, streamConnectionStatusLabel } from './monitoring-utils.js';
import { formatTimestamp } from './reporting-utils.js';
import { useTerminalHeartbeatStream } from './useTerminalHeartbeatStream.js';

type TerminalHeartbeatEventsPanelProps = {
  storeId: string;
  terminalId?: string;
  maxEvents?: number;
  title?: string;
};

function TerminalHeartbeatEventsPanelContent({
  storeId,
  terminalId,
  maxEvents,
  title = 'Terminal events',
}: TerminalHeartbeatEventsPanelProps) {
  const { events, connectionStatus } = useTerminalHeartbeatStream(storeId, {
    terminalId,
    maxEvents,
  });

  return (
    <div className="panel">
      <div className="panel-heading">
        <div>
          <h3>{title}</h3>
          <p className="muted">Showing terminal_heartbeat events only.</p>
        </div>
        <span className={streamConnectionStatusClass(connectionStatus)}>
          {streamConnectionStatusLabel(connectionStatus)}
        </span>
      </div>

      {storeId.length === 0 ? (
        <p className="muted">Select a store to subscribe to terminal events.</p>
      ) : connectionStatus === 'connected' && events.length === 0 ? (
        <p className="muted">Waiting for terminal heartbeats…</p>
      ) : events.length > 0 ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Received</th>
                <th>Terminal</th>
                <th>Kind</th>
                <th>Status</th>
                <th>Last seen</th>
              </tr>
            </thead>
            <tbody>
              {events.map((event) => (
                <tr key={`${event.receivedAt}-${event.terminalId}-${event.updatedAt}`}>
                  <td>{formatTimestamp(event.receivedAt)}</td>
                  <td>
                    <Link to={terminalMonitoringHref(storeId, event.terminalId)}>
                      {event.terminalId}
                    </Link>
                  </td>
                  <td>{event.kind}</td>
                  <td>{event.status}</td>
                  <td>{formatTimestamp(event.lastSeenAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : connectionStatus === 'error' ? (
        <p className="muted">Unable to connect to the terminal event stream.</p>
      ) : (
        <p className="muted">Connecting to terminal event stream…</p>
      )}
    </div>
  );
}

export function TerminalHeartbeatEventsPanel(props: TerminalHeartbeatEventsPanelProps) {
  const streamKey = `${props.storeId}:${props.terminalId ?? ''}:${props.maxEvents ?? 50}`;

  return <TerminalHeartbeatEventsPanelContent key={streamKey} {...props} />;
}
