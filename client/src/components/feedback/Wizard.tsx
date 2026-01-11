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

  return createPortal(
    <div className={`fixed inset-0 z-[9999] flex transition-opacity duration-300 ${animate && !closing ? 'opacity-100' : 'opacity-0'}`}>
      <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={onClose} />
      
      <div className={`relative m-auto w-full max-w-3xl bg-neutral-900 rounded-2xl shadow-2xl border border-neutral-800 overflow-hidden transition-all duration-300 ${animate && !closing ? 'scale-100 translate-y-0' : 'scale-95 translate-y-4'}`}>
        <div className="flex items-center justify-between px-6 py-4 border-b border-neutral-800">
          <div className="flex items-center gap-2 sm:gap-4 overflow-x-auto">
            {steps.map((s, i) => (
              <button
                key={i}
                type="button"
                onClick={() => i < currentStep && onStepChange(i)}
                disabled={i > currentStep}
                className={`flex items-center gap-1.5 transition-colors flex-shrink-0 ${
                  i === currentStep 
                    ? 'text-neutral-100' 
                    : i < currentStep 
                      ? 'text-neutral-400 hover:text-neutral-200 cursor-pointer' 
                      : 'text-neutral-600 cursor-not-allowed'
                }`}
              >
                <span className={`flex items-center justify-center w-6 h-6 rounded-full text-xs font-semibold transition-colors ${
                  i === currentStep 
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
            className="p-2 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 transition-colors flex-shrink-0 ml-2"
          >
            <Icons.x className="w-5 h-5" />
          </button>
        </div>

        {headerContent && <div className="px-6 pt-6">{headerContent}</div>}

        <div className="px-6 py-8 min-h-[320px] max-h-[60vh] overflow-y-auto">
          {children}
        </div>

        <div className="flex items-center justify-between px-6 py-4 bg-neutral-900/50 border-t border-neutral-800">
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
