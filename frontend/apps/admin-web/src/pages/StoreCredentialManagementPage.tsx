import { useListStores } from '@mercadia/api-clients-central';
import {
  addActorCredentialBinding,
  clearSessionToken as clearStoreEdgeSessionToken,
  createAuthSession,
  getSessionToken as getStoreEdgeSessionToken,
  revokeActorCredentialBinding,
  setActorCredentialPolicy,
  setSessionToken as setStoreEdgeSessionToken,
  setStoreCredentialPolicy,
  useGetCredentialManagement,
  type AddActorCredentialBindingBody,
  type GetCredentialManagement200ActorsItem,
  type GetCredentialManagement200ActorsItemCredentialBindingsItem,
} from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { CheckboxField, SelectField, TextField } from '@/components/FormControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';

type CredentialKind = AddActorCredentialBindingBody['kind'];
type CredentialPolicyDraft = {
  required: boolean;
  allowedKinds: CredentialKind[];
};
type ActorPolicyDraft = CredentialPolicyDraft & {
  inheritStorePolicy: boolean;
};

const CREDENTIAL_KINDS: CredentialKind[] = ['ibutton', 'msr_card', 'barcode_card'];

function toggleKind(
  kinds: CredentialKind[],
  kind: CredentialKind,
  checked: boolean,
): CredentialKind[] {
  if (checked) {
    return [...new Set([...kinds, kind])];
  }
  return kinds.filter((candidate) => candidate !== kind);
}

function kindLabelKey(kind: CredentialKind): string {
  return `credentials.kinds.${kind}`;
}

function bindingLabel(binding: GetCredentialManagement200ActorsItemCredentialBindingsItem): string {
  return binding.maskedToken || binding.tokenFingerprint;
}

function actorPolicyDraftFromActor(
  actor: GetCredentialManagement200ActorsItem | null,
): ActorPolicyDraft {
  return {
    inheritStorePolicy: !actor?.credentialPolicy,
    required: actor?.credentialPolicy?.required ?? false,
    allowedKinds: (actor?.credentialPolicy?.allowedKinds ?? []) as CredentialKind[],
  };
}

export function StoreCredentialManagementPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [managerActorId, setManagerActorId] = useState('admin-1');
  const [managerPin, setManagerPin] = useState('');
  const [managerCredentialKind, setManagerCredentialKind] = useState<CredentialKind>('ibutton');
  const [managerCredentialToken, setManagerCredentialToken] = useState('');
  const [hasManagerSession, setHasManagerSession] = useState(
    () => getStoreEdgeSessionToken() !== null,
  );
  const [storePolicyDraft, setStorePolicyDraft] = useState<CredentialPolicyDraft | null>(null);
  const [selectedActorId, setSelectedActorId] = useState('');
  const [actorPolicyDraft, setActorPolicyDraft] = useState<ActorPolicyDraft | null>(null);
  const [bindingKind, setBindingKind] = useState<CredentialKind>('ibutton');
  const [bindingToken, setBindingToken] = useState('');
  const [bindingMaskedToken, setBindingMaskedToken] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const credentialsQuery = useGetCredentialManagement(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 && hasManagerSession },
  });
  const credentials =
    hasManagerSession && credentialsQuery.data?.status === 200 ? credentialsQuery.data.data : null;
  const actors = credentials?.actors ?? [];
  const selectedActor = actors.find((actor) => actor.id === selectedActorId) ?? actors[0] ?? null;
  const targetActorId = selectedActor?.id ?? '';
  const storePolicyForm = storePolicyDraft ?? {
    required: credentials?.storePolicy.required ?? true,
    allowedKinds: (credentials?.storePolicy.allowedKinds ?? CREDENTIAL_KINDS) as CredentialKind[],
  };
  const actorPolicyForm = actorPolicyDraft ?? actorPolicyDraftFromActor(selectedActor);
  const loadError =
    hasManagerSession && credentialsQuery.error != null
      ? getApiErrorMessage(credentialsQuery.error)
      : null;

  async function loginManager(): Promise<void> {
    setIsSubmitting(true);
    setError('');
    setMessage('');
    try {
      const credentialToken = managerCredentialToken.trim();
      const response = await createAuthSession({
        actorId: managerActorId.trim(),
        pin: managerPin.trim(),
        storeId: activeStoreId,
        credentialFactor:
          credentialToken.length > 0
            ? { kind: managerCredentialKind, token: credentialToken }
            : undefined,
      });
      if (response.status === 201) {
        setStoreEdgeSessionToken(response.data.session.token);
        setHasManagerSession(true);
        setManagerPin('');
        setManagerCredentialToken('');
        setMessage(t('credentials.managerLoggedIn'));
      }
    } catch (err) {
      setError(getApiErrorMessage(err));
      setHasManagerSession(false);
    } finally {
      setIsSubmitting(false);
    }
  }

  async function runCommand(action: () => Promise<unknown>, successKey: string): Promise<boolean> {
    setIsSubmitting(true);
    setError('');
    setMessage('');
    try {
      await action();
      await credentialsQuery.refetch();
      setMessage(t(successKey));
      return true;
    } catch (err) {
      setError(getApiErrorMessage(err));
      return false;
    } finally {
      setIsSubmitting(false);
    }
  }

  function applyActorSelection(actor: GetCredentialManagement200ActorsItem): void {
    setSelectedActorId(actor.id);
    setActorPolicyDraft(null);
  }

  const canSubmitStorePolicy = !storePolicyForm.required || storePolicyForm.allowedKinds.length > 0;
  const canSubmitActorPolicy =
    actorPolicyForm.inheritStorePolicy ||
    !actorPolicyForm.required ||
    actorPolicyForm.allowedKinds.length > 0;
  const canSubmitBinding = targetActorId.length > 0 && bindingToken.trim().length > 0;

  return (
    <section className="stack">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('credentials.title')}</h2>
            <p className="muted">{t('credentials.subtitle')}</p>
          </div>
          <div className="header-actions-inline">
            <Button
              type="button"
              variant="secondary"
              disabled={
                credentialsQuery.isFetching || activeStoreId.length === 0 || !hasManagerSession
              }
              onClick={() => void credentialsQuery.refetch()}
            >
              {credentialsQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
            </Button>
          </div>
        </div>

        <div className="form-grid form-grid--two">
          <StorePicker
            stores={stores}
            value={activeStoreId}
            onChange={(storeId) => {
              setSelectedStoreId(storeId);
              setStorePolicyDraft(null);
              setSelectedActorId('');
              setActorPolicyDraft(null);
              setHasManagerSession(false);
              setManagerPin('');
              setManagerCredentialToken('');
              setBindingToken('');
              setBindingMaskedToken('');
              clearStoreEdgeSessionToken();
              setMessage('');
              setError('');
            }}
          />
          <TextField
            label={t('credentials.managerActorId')}
            value={managerActorId}
            onValueChange={setManagerActorId}
            placeholder={t('credentials.managerActorPlaceholder')}
          />
          <TextField
            label={t('credentials.managerPin')}
            type="password"
            autoComplete="off"
            value={managerPin}
            onValueChange={setManagerPin}
            placeholder={t('credentials.managerPinPlaceholder')}
          />
          <SelectField
            label={t('credentials.managerCredentialKind')}
            value={managerCredentialKind}
            onValueChange={(value) => setManagerCredentialKind(value as CredentialKind)}
          >
            {CREDENTIAL_KINDS.map((kind) => (
              <option key={kind} value={kind}>
                {t(kindLabelKey(kind))}
              </option>
            ))}
          </SelectField>
          <TextField
            label={t('credentials.managerCredentialTokenOptional')}
            type="password"
            autoComplete="off"
            spellCheck={false}
            value={managerCredentialToken}
            onValueChange={setManagerCredentialToken}
            placeholder={t('credentials.rawTokenPlaceholder')}
          />
        </div>

        <div className="header-actions-inline">
          <Button
            type="button"
            disabled={
              isSubmitting ||
              activeStoreId.length === 0 ||
              managerActorId.trim().length === 0 ||
              managerPin.trim().length === 0
            }
            onClick={() => void loginManager()}
          >
            {isSubmitting ? t('common.submitting') : t('credentials.managerLogin')}
          </Button>
          {hasManagerSession && (
            <span className="muted">{t('credentials.managerSessionReady')}</span>
          )}
        </div>
        <p className="muted">{t('credentials.managerLoginHint')}</p>

        {loadError && <p className="error">{loadError}</p>}
        {!activeStoreId && <p className="muted">{t('common.selectStore')}</p>}
        {message && <p className="muted">{message}</p>}
        {error && <p className="error">{error}</p>}
      </div>

      {credentials && (
        <>
          <div className="panel">
            <h3>{t('credentials.storePolicy')}</h3>
            <p className="muted">
              {t('credentials.currentStorePolicy', {
                required: credentials.storePolicy.required ? t('common.yes') : t('common.no'),
                kinds: credentials.storePolicy.allowedKinds
                  .map((kind) => t(kindLabelKey(kind as CredentialKind)))
                  .join(', '),
              })}
            </p>
            <div className="stack stack--compact">
              <CheckboxField
                checked={storePolicyForm.required}
                label={t('credentials.required')}
                onCheckedChange={(required) =>
                  setStorePolicyDraft({ ...storePolicyForm, required })
                }
              />
              <fieldset className="role-fieldset">
                <legend>{t('credentials.allowedKinds')}</legend>
                <div className="role-options">
                  {CREDENTIAL_KINDS.map((kind) => (
                    <CheckboxField
                      key={kind}
                      checked={storePolicyForm.allowedKinds.includes(kind)}
                      label={t(kindLabelKey(kind))}
                      onCheckedChange={(checked) =>
                        setStorePolicyDraft({
                          ...storePolicyForm,
                          allowedKinds: toggleKind(storePolicyForm.allowedKinds, kind, checked),
                        })
                      }
                    />
                  ))}
                </div>
              </fieldset>
              <Button
                type="button"
                disabled={isSubmitting || !canSubmitStorePolicy}
                onClick={() =>
                  void runCommand(
                    () =>
                      setStoreCredentialPolicy(
                        activeStoreId,
                        {
                          required: storePolicyForm.required,
                          allowedKinds: storePolicyForm.allowedKinds,
                        },
                        { headers: createIdempotencyHeaders() },
                      ),
                    'credentials.storePolicySaved',
                  )
                }
              >
                {isSubmitting ? t('common.submitting') : t('credentials.saveStorePolicy')}
              </Button>
            </div>
          </div>

          <div className="panel">
            <h3>{t('credentials.actorPolicies')}</h3>
            <div className="form-grid form-grid--two">
              <SelectField
                label={t('credentials.actor')}
                value={targetActorId}
                onValueChange={(id) => {
                  const actor = actors.find((candidate) => candidate.id === id);
                  if (actor) applyActorSelection(actor);
                }}
              >
                {actors.map((actor) => (
                  <option key={actor.id} value={actor.id}>
                    {actor.id} ({actor.roles.join(', ')})
                  </option>
                ))}
              </SelectField>
              <CheckboxField
                checked={actorPolicyForm.inheritStorePolicy}
                label={t('credentials.inheritStorePolicy')}
                onCheckedChange={(inheritStorePolicy) =>
                  setActorPolicyDraft({ ...actorPolicyForm, inheritStorePolicy })
                }
              />
            </div>
            {!actorPolicyForm.inheritStorePolicy && (
              <div className="stack stack--compact">
                <CheckboxField
                  checked={actorPolicyForm.required}
                  label={t('credentials.required')}
                  onCheckedChange={(required) =>
                    setActorPolicyDraft({ ...actorPolicyForm, required })
                  }
                />
                <fieldset className="role-fieldset">
                  <legend>{t('credentials.allowedKinds')}</legend>
                  <div className="role-options">
                    {CREDENTIAL_KINDS.map((kind) => (
                      <CheckboxField
                        key={kind}
                        checked={actorPolicyForm.allowedKinds.includes(kind)}
                        label={t(kindLabelKey(kind))}
                        onCheckedChange={(checked) =>
                          setActorPolicyDraft({
                            ...actorPolicyForm,
                            allowedKinds: toggleKind(actorPolicyForm.allowedKinds, kind, checked),
                          })
                        }
                      />
                    ))}
                  </div>
                </fieldset>
              </div>
            )}
            <Button
              type="button"
              disabled={isSubmitting || !canSubmitActorPolicy || targetActorId.length === 0}
              onClick={() =>
                void runCommand(
                  () =>
                    setActorCredentialPolicy(
                      activeStoreId,
                      targetActorId,
                      {
                        inheritStorePolicy: actorPolicyForm.inheritStorePolicy,
                        required: actorPolicyForm.required,
                        allowedKinds: actorPolicyForm.allowedKinds,
                      },
                      { headers: createIdempotencyHeaders() },
                    ),
                  'credentials.actorPolicySaved',
                )
              }
            >
              {isSubmitting ? t('common.submitting') : t('credentials.saveActorPolicy')}
            </Button>
          </div>

          <div className="panel">
            <h3>{t('credentials.bindings')}</h3>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>{t('credentials.actor')}</th>
                    <th>{t('credentials.kind')}</th>
                    <th>{t('credentials.maskedToken')}</th>
                    <th>{t('credentials.fingerprint')}</th>
                    <th>{t('credentials.active')}</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {actors.flatMap((actor) =>
                    actor.credentialBindings.map((binding) => (
                      <tr key={`${actor.id}-${binding.kind}-${binding.tokenFingerprint}`}>
                        <td>{actor.id}</td>
                        <td>{t(kindLabelKey(binding.kind as CredentialKind))}</td>
                        <td>{bindingLabel(binding)}</td>
                        <td>{binding.tokenFingerprint}</td>
                        <td>{binding.active ? t('common.yes') : t('common.no')}</td>
                        <td>
                          <Button
                            type="button"
                            size="sm"
                            variant="secondary"
                            disabled={isSubmitting || !binding.active}
                            onClick={() =>
                              void runCommand(
                                () =>
                                  revokeActorCredentialBinding(
                                    activeStoreId,
                                    actor.id,
                                    {
                                      kind: binding.kind,
                                      tokenFingerprint: binding.tokenFingerprint,
                                    },
                                    { headers: createIdempotencyHeaders() },
                                  ),
                                'credentials.bindingRevoked',
                              )
                            }
                          >
                            {t('credentials.revoke')}
                          </Button>
                        </td>
                      </tr>
                    )),
                  )}
                </tbody>
              </table>
            </div>

            <div className="form-grid form-grid--two">
              <SelectField
                label={t('credentials.kind')}
                value={bindingKind}
                onValueChange={(value) => setBindingKind(value as CredentialKind)}
              >
                {CREDENTIAL_KINDS.map((kind) => (
                  <option key={kind} value={kind}>
                    {t(kindLabelKey(kind))}
                  </option>
                ))}
              </SelectField>
              <TextField
                label={t('credentials.maskedTokenOptional')}
                value={bindingMaskedToken}
                onValueChange={setBindingMaskedToken}
                placeholder={t('credentials.maskedTokenPlaceholder')}
              />
              <TextField
                label={t('credentials.rawToken')}
                type="password"
                autoComplete="off"
                spellCheck={false}
                value={bindingToken}
                onValueChange={setBindingToken}
                placeholder={t('credentials.rawTokenPlaceholder')}
              />
            </div>
            <p className="muted">{t('credentials.rawTokenHint')}</p>
            <Button
              type="button"
              disabled={isSubmitting || !canSubmitBinding}
              onClick={() =>
                void runCommand(
                  () =>
                    addActorCredentialBinding(
                      activeStoreId,
                      targetActorId,
                      {
                        kind: bindingKind,
                        token: bindingToken.trim(),
                        maskedToken: bindingMaskedToken || undefined,
                      },
                      { headers: createIdempotencyHeaders() },
                    ),
                  'credentials.bindingAdded',
                ).then((ok) => {
                  if (!ok) return;
                  setBindingToken('');
                  setBindingMaskedToken('');
                })
              }
            >
              {isSubmitting ? t('common.submitting') : t('credentials.addBinding')}
            </Button>
          </div>
        </>
      )}
    </section>
  );
}
