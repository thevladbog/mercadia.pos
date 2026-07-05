const PIN_DISPLAY_BOXES = 6;

/**
 * Masked PIN display: a fixed row of boxes, one filled dot per entered
 * digit and a blinking caret on the next empty box (design screen 01b).
 * Deliberately never renders the raw PIN characters — see `PinStep.tsx`
 * for why this exists alongside `Numpad`'s own (now-hidden) display.
 *
 * The PIN itself has no length limit enforced anywhere in the API contract
 * (see plan 020's "Current state"), so this is purely a fixed-size visual:
 * once `length` reaches `PIN_DISPLAY_BOXES`, every box shows a dot and the
 * caret is simply omitted (no 7th box is added).
 */
export function PinDots({ length }: { length: number }) {
  return (
    <div className="sr-login-pin-dots" aria-hidden="true">
      {Array.from({ length: PIN_DISPLAY_BOXES }, (_, index) => {
        const filled = index < length;
        const isCaret = index === length;
        const className = [
          'sr-login-pin-box',
          filled && 'sr-login-pin-box--filled',
          isCaret && 'sr-login-pin-box--caret',
        ]
          .filter(Boolean)
          .join(' ');
        return (
          <div key={index} className={className}>
            {filled && <span className="sr-login-pin-dot" />}
            {isCaret && <span className="sr-login-pin-caret" />}
          </div>
        );
      })}
    </div>
  );
}
