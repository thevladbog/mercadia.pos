import type { Decorator, Preview } from '@storybook/react-vite';

import '../src/styles/index.css';
import '../src/stories/storybook.css';
import { ThemeProvider } from '../src/theme/ThemeProvider.js';
import type { AccentPreset, ColorMode, Surface } from '../src/theme/types.js';

const withMercadiaTheme: Decorator = (Story, context) => {
  const surface = (context.globals.surface ?? 'admin') as Surface;
  const colorMode = (context.globals.colorMode ?? 'light') as ColorMode;
  const accentPreset = (context.globals.accentPreset ?? 'neutral') as AccentPreset;

  return (
    <ThemeProvider
      key={`${surface}-${colorMode}-${accentPreset}`}
      defaultTheme={{ surface, colorMode, accentPreset }}
    >
      <div className="mercadia-storybook-shell">
        <Story />
      </div>
    </ThemeProvider>
  );
};

const preview: Preview = {
  decorators: [withMercadiaTheme],
  globalTypes: {
    surface: {
      description: 'Mercadia UI surface preset',
      toolbar: {
        title: 'Surface',
        icon: 'browser',
        items: ['admin', 'terminal', 'sco', 'senior-cashier'],
        dynamicTitle: true,
      },
    },
    colorMode: {
      description: 'Mercadia UI color mode',
      toolbar: {
        title: 'Mode',
        icon: 'mirror',
        items: ['light', 'dark'],
        dynamicTitle: true,
      },
    },
    accentPreset: {
      description: 'Mercadia UI accent preset',
      toolbar: {
        title: 'Accent',
        icon: 'paintbrush',
        items: ['neutral', 'sale', 'return', 'sco'],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: {
    surface: 'admin',
    colorMode: 'light',
    accentPreset: 'neutral',
  },
  parameters: {
    controls: {
      expanded: true,
    },
    backgrounds: {
      disable: true,
    },
  },
};

export default preview;
