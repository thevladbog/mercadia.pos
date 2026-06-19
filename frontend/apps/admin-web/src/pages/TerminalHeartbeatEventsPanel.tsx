import { useTranslation } from 'react-i18next';
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
  title,
}: TerminalHeartbeatEventsPanelProps) {
  const { t } = useTranslation();
  const { events, connectionStatus } = useTerminalHeartbeatStream(storeId, {
    terminalId,
    maxEvents,
  });

  const panelTitle = title ?? t('monitoring.eventsTitle');

  return (
    <div className="panel">
      <div className="panel-heading">
        <div>
          <h3>{panelTitle}</h3>
          <p className="muted">{t('monitoring.eventsNote')}</p>
        </div>
        <span className={streamConnectionStatusClass(connectionStatus)}>
          {streamConnectionStatusLabel(connectionStatus)}
        </span>
      </div>

      {storeId.length === 0 ? (
        <p className="muted">{t('monitoring.subscribeStore')}</p>
      ) : connectionStatus === 'connected' && events.length === 0 ? (
        <p className="muted">{t('monitoring.waitingHeartbeats')}</p>
      ) : events.length > 0 ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>{t('monitoring.eventReceived')}</th>
                <th>{t('monitoring.terminal')}</th>
                <th>{t('monitoring.kind')}</th>
                <th>{t('monitoring.status')}</th>
                <th>{t('monitoring.lastSeen')}</th>
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
        <p className="muted">{t('monitoring.streamError')}</p>
      ) : (
        <p className="muted">{t('monitoring.connectingStream')}</p>
      )}
    </div>
  );
}

export function TerminalHeartbeatEventsPanel(props: TerminalHeartbeatEventsPanelProps) {
  const streamKey = `${props.storeId}:${props.terminalId ?? ''}:${props.maxEvents ?? 50}`;

  return <TerminalHeartbeatEventsPanelContent key={streamKey} {...props} />;
}
