import { ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Icons } from '../Icons';
import Button from '../ui/Button';

interface WizardStep {
  label: string;
  name: string;
}

interface WizardProps {
  steps: WizardStep[];
  currentStep: number;
  onStepChange: (step: number) => void;
  onClose: () => void;
  onComplete: () => void;
  canProceed: boolean;
  canFinish?: boolean;
  loading?: boolean;
  children: ReactNode;
  completeLabel?: string;
  headerContent?: ReactNode;
  animate: boolean;
  closing: boolean;
}

export default function Wizard({
  steps,
  currentStep,
  onStepChange,
  onClose,
  onComplete,
  canProceed,
  canFinish,
  loading,
  children,
  completeLabel = 'Complete',
  headerContent,
  animate,
  closing,
}: WizardProps) {
  const isLastStep = currentStep === steps.length - 1;
  const isActive = animate && !closing;

  return createPortal(
    <div className="fixed inset-0 z-[9999] flex justify-end">
      <div
        className={`absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity duration-300 ${isActive ? 'opacity-100' : 'opacity-0'}`}
        onClick={onClose}
      />

      <div className={`
        relative w-full max-w-3xl h-full flex flex-col
        bg-neutral-900/90 backdrop-blur-2xl
        border-l border-neutral-700/40
        shadow-2xl shadow-black/50
        transition-transform duration-300 ease-out
        ${isActive ? 'translate-x-0' : 'translate-x-full'}
      `}>
        <div className="absolute inset-y-0 -left-px w-px bg-gradient-to-b from-transparent via-neutral-500/40 to-transparent" />

        <div className="shrink-0 flex items-center justify-between px-6 py-4 border-b border-neutral-700/40">
          <div className="flex items-center gap-2 sm:gap-4 overflow-x-auto">
            {steps.map((s, i) => (
              <button
                key={i}
                type="button"
                onClick={() => i < currentStep && onStepChange(i)}
                disabled={i > currentStep}
                className={`flex items-center gap-1.5 transition-colors flex-shrink-0 ${i === currentStep
                  ? 'text-neutral-100'
                  : i < currentStep
                    ? 'text-neutral-400 hover:text-neutral-200 cursor-pointer'
                    : 'text-neutral-600 cursor-not-allowed'
                  }`}
              >
                <span className={`flex items-center justify-center w-6 h-6 rounded-full text-xs font-semibold transition-colors ${i === currentStep
                  ? 'bg-white text-black'
                  : i < currentStep
                    ? 'bg-neutral-700 text-neutral-300'
                    : 'bg-neutral-800 text-neutral-600'
                  }`}>
                  {i < currentStep ? <Icons.check className="w-3 h-3" /> : i + 1}
                </span>
                <span className="text-xs font-medium hidden md:block">{s.name}</span>
              </button>
            ))}
          </div>
          <button
            onClick={onClose}
            className="shrink-0 ml-2 inline-flex h-7 w-7 items-center justify-center rounded-lg text-neutral-500 hover:text-neutral-200 hover:bg-neutral-700/50 transition-all duration-150 cursor-pointer"
            aria-label="Close"
          >
            <Icons.xFilled className="w-4 h-4" />
          </button>
        </div>

        {headerContent && <div className="px-6 pt-6">{headerContent}</div>}

        <div className="flex-1 overflow-y-auto px-6 py-8">
          {children}
        </div>

        <div className="shrink-0 flex items-center justify-between px-6 py-4 border-t border-neutral-700/40 bg-neutral-900/50">
          <button
            onClick={() => currentStep > 0 ? onStepChange(currentStep - 1) : onClose()}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-neutral-400 hover:text-neutral-100 rounded-lg transition-colors"
          >
            <Icons.chevronLeft className="w-4 h-4" />
            {currentStep === 0 ? 'Cancel' : 'Back'}
          </button>

          <div className="flex items-center gap-3">
            {canFinish && !isLastStep && (
              <Button onClick={onComplete} disabled={loading} loading={loading} variant="secondary">
                Finish
              </Button>
            )}
            {isLastStep ? (
              <Button onClick={onComplete} disabled={!canProceed || loading} loading={loading}>
                {completeLabel}
              </Button>
            ) : (
              <Button
                onClick={() => onStepChange(currentStep + 1)}
                disabled={!canProceed}
              >
                Continue
                <Icons.chevronRight className="w-4 h-4" />
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
