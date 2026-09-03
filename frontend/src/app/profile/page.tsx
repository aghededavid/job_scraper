'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  User, Mail, Bell, Globe, DollarSign, Briefcase, RefreshCw, CheckCircle2, AlertCircle
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { user as userApi } from '@/lib/api';
import { cn } from '@/lib/utils';
import type { UserPreferences } from '@/lib/types';

const CATEGORY_OPTIONS = [
  { slug: 'frontend', name: 'Frontend', icon: '🖥️' },
  { slug: 'backend', name: 'Backend', icon: '⚙️' },
  { slug: 'full-stack', name: 'Full Stack', icon: '🔗' },
  { slug: 'devops-cloud', name: 'DevOps & Cloud', icon: '☁️' },
  { slug: 'ai-ml', name: 'AI & ML', icon: '🤖' },
  { slug: 'data-science', name: 'Data Science', icon: '📊' },
  { slug: 'mobile', name: 'Mobile', icon: '📱' },
  { slug: 'ui-ux-design', name: 'UI/UX Design', icon: '🎨' },
  { slug: 'cybersecurity', name: 'Cybersecurity', icon: '🔒' },
  { slug: 'product-management', name: 'Product', icon: '📋' },
];

const REMOTE_OPTIONS = [
  { value: 'fully-remote', label: '🌍 Fully Remote' },
  { value: 'remote-friendly', label: '🔀 Remote-Friendly' },
  { value: 'hybrid', label: '🏢 Hybrid' },
];

const EXPERIENCE_OPTIONS = [
  { value: 'entry', label: 'Entry Level' },
  { value: 'mid', label: 'Mid Level' },
  { value: 'senior', label: 'Senior' },
  { value: 'lead', label: 'Lead / Staff' },
  { value: 'executive', label: 'Executive' },
];

const SKILL_SUGGESTIONS = [
  'TypeScript', 'Python', 'React', 'Node.js', 'Go', 'Rust',
  'AWS', 'Docker', 'Kubernetes', 'PostgreSQL', 'GraphQL', 'Next.js',
  'Vue.js', 'Terraform', 'Redis', 'MongoDB', 'Swift', 'Kotlin',
];

function MultiSelect<T extends string>({
  options,
  selected,
  onChange,
  renderOption,
}: {
  options: { value: T; label: string }[];
  selected: T[];
  onChange: (val: T[]) => void;
  renderOption?: (opt: { value: T; label: string }) => React.ReactNode;
}) {
  const toggle = (v: T) =>
    onChange(selected.includes(v) ? selected.filter((s) => s !== v) : [...selected, v]);
  return (
    <div className="flex flex-wrap gap-2">
      {options.map((opt) => {
        const active = selected.includes(opt.value);
        return (
          <button
            key={opt.value}
            type="button"
            onClick={() => toggle(opt.value)}
            className={cn(
              'px-3 py-1.5 rounded-lg text-sm font-medium border transition-all',
              active
                ? 'bg-brand-500/20 text-brand-300 border-brand-500/40'
                : 'text-surface-300 border-white/[0.07] hover:border-white/15 hover:text-white bg-surface-700/40'
            )}
          >
            {renderOption ? renderOption(opt) : opt.label}
          </button>
        );
      })}
    </div>
  );
}

export default function ProfilePage() {
  const { user, isLoading, isAuthenticated } = useAuth();
  const router = useRouter();

  const [name, setName] = useState('');
  const [prefs, setPrefs] = useState<Partial<UserPreferences>>({});
  const [skillInput, setSkillInput] = useState('');
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (!isLoading && !isAuthenticated) router.replace('/auth/login');
  }, [isLoading, isAuthenticated, router]);

  useEffect(() => {
    if (user) {
      setName(user.name || '');
      userApi.me().then((res) => {
        // Preferences come as part of the user object if expanded
      });
    }
  }, [user]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setStatus('idle');
    try {
      const [profileRes, prefsRes] = await Promise.all([
        userApi.update({ name }),
        userApi.updatePreferences(prefs),
      ]);
      if (profileRes.success && prefsRes.success) {
        setStatus('success');
        setTimeout(() => setStatus('idle'), 3000);
      } else {
        setStatus('error');
        setErrorMsg('Failed to save changes');
      }
    } catch {
      setStatus('error');
      setErrorMsg('An error occurred');
    } finally {
      setSaving(false);
    }
  };

  const addSkill = () => {
    const s = skillInput.trim();
    if (!s) return;
    setPrefs((p) => ({ ...p, preferred_skills: [...(p.preferred_skills ?? []), s] }));
    setSkillInput('');
  };

  const removeSkill = (skill: string) => {
    setPrefs((p) => ({ ...p, preferred_skills: (p.preferred_skills ?? []).filter((s) => s !== skill) }));
  };

  if (isLoading || !isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <RefreshCw className="w-6 h-6 animate-spin text-brand-400" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-glow-top">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-white mb-1">Profile & Preferences</h1>
          <p className="text-surface-400">Customize your job discovery experience</p>
        </div>

        <form onSubmit={handleSave} className="space-y-6">
          {/* Status alerts */}
          {status === 'success' && (
            <div className="flex items-center gap-2 p-4 rounded-xl bg-accent-emerald/10 border border-accent-emerald/30 text-accent-emerald text-sm">
              <CheckCircle2 className="w-4 h-4 flex-shrink-0" />
              Preferences saved successfully!
            </div>
          )}
          {status === 'error' && (
            <div className="flex items-center gap-2 p-4 rounded-xl bg-accent-rose/10 border border-accent-rose/30 text-accent-rose text-sm">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              {errorMsg}
            </div>
          )}

          {/* Account info */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
              <User className="w-4 h-4 text-brand-400" />
              Account Information
            </h2>
            <div className="space-y-4">
              <div>
                <label htmlFor="profile-name" className="block text-sm font-medium text-surface-300 mb-1.5">
                  Full name
                </label>
                <input
                  id="profile-name"
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="input"
                  placeholder="Your name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-surface-300 mb-1.5">
                  Email address
                </label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-400" />
                  <input
                    type="email"
                    value={user?.email || ''}
                    disabled
                    className="input pl-10 opacity-60 cursor-not-allowed"
                  />
                </div>
                <p className="text-xs text-surface-500 mt-1">Email cannot be changed</p>
              </div>
            </div>
          </section>

          {/* Categories */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-2 flex items-center gap-2">
              <Briefcase className="w-4 h-4 text-brand-400" />
              Preferred Categories
            </h2>
            <p className="text-xs text-surface-400 mb-4">Select the types of roles you&apos;re interested in</p>
            <div className="flex flex-wrap gap-2">
              {CATEGORY_OPTIONS.map((cat) => {
                const active = (prefs.preferred_categories ?? []).includes(cat.slug);
                return (
                  <button
                    key={cat.slug}
                    type="button"
                    onClick={() => setPrefs((p) => ({
                      ...p,
                      preferred_categories: active
                        ? (p.preferred_categories ?? []).filter((c) => c !== cat.slug)
                        : [...(p.preferred_categories ?? []), cat.slug],
                    }))}
                    className={cn(
                      'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium border transition-all',
                      active
                        ? 'bg-brand-500/20 text-brand-300 border-brand-500/40'
                        : 'text-surface-300 border-white/[0.07] hover:border-white/15 hover:text-white bg-surface-700/40'
                    )}
                  >
                    <span>{cat.icon}</span>
                    {cat.name}
                  </button>
                );
              })}
            </div>
          </section>

          {/* Skills */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-2 flex items-center gap-2">
              <Globe className="w-4 h-4 text-brand-400" />
              Preferred Skills
            </h2>
            <p className="text-xs text-surface-400 mb-4">Jobs matching these skills will be highlighted</p>

            {/* Selected skills */}
            {(prefs.preferred_skills ?? []).length > 0 && (
              <div className="flex flex-wrap gap-2 mb-4">
                {(prefs.preferred_skills ?? []).map((skill) => (
                  <span key={skill} className="tag-brand flex items-center gap-1">
                    {skill}
                    <button
                      type="button"
                      onClick={() => removeSkill(skill)}
                      className="ml-0.5 text-brand-400/60 hover:text-accent-rose transition-colors"
                      aria-label={`Remove ${skill}`}
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            )}

            {/* Skill input */}
            <div className="flex gap-2 mb-3">
              <input
                type="text"
                value={skillInput}
                onChange={(e) => setSkillInput(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addSkill(); } }}
                placeholder="Add a skill (press Enter)"
                className="input flex-1 text-sm py-2"
              />
              <button type="button" onClick={addSkill} className="btn-ghost text-sm px-4">
                Add
              </button>
            </div>

            {/* Suggestions */}
            <div className="flex flex-wrap gap-1.5">
              {SKILL_SUGGESTIONS.filter((s) => !(prefs.preferred_skills ?? []).includes(s)).slice(0, 10).map((skill) => (
                <button
                  key={skill}
                  type="button"
                  onClick={() => setPrefs((p) => ({ ...p, preferred_skills: [...(p.preferred_skills ?? []), skill] }))}
                  className="px-2.5 py-1 rounded-lg text-xs text-surface-400 border border-white/[0.06] hover:border-brand-500/40 hover:text-brand-300 hover:bg-brand-950/30 transition-all"
                >
                  + {skill}
                </button>
              ))}
            </div>
          </section>

          {/* Experience & Remote */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
              <Briefcase className="w-4 h-4 text-brand-400" />
              Work Preferences
            </h2>
            <div className="space-y-5">
              <div>
                <p className="text-sm font-medium text-surface-300 mb-2">Experience Level</p>
                <MultiSelect
                  options={EXPERIENCE_OPTIONS as { value: string; label: string }[]}
                  selected={(prefs.preferred_experience_levels ?? []) as string[]}
                  onChange={(v) => setPrefs((p) => ({ ...p, preferred_experience_levels: v as UserPreferences['preferred_experience_levels'] }))}
                />
              </div>
              <div>
                <p className="text-sm font-medium text-surface-300 mb-2">Remote Type</p>
                <MultiSelect
                  options={REMOTE_OPTIONS as { value: string; label: string }[]}
                  selected={(prefs.preferred_remote_types ?? []) as string[]}
                  onChange={(v) => setPrefs((p) => ({ ...p, preferred_remote_types: v as UserPreferences['preferred_remote_types'] }))}
                />
              </div>
            </div>
          </section>

          {/* Salary */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
              <DollarSign className="w-4 h-4 text-brand-400" />
              Salary Expectations (USD/yr)
            </h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="salary-min" className="block text-sm text-surface-400 mb-1.5">Minimum</label>
                <input
                  id="salary-min"
                  type="number"
                  value={prefs.salary_min ?? ''}
                  onChange={(e) => setPrefs((p) => ({ ...p, salary_min: e.target.value ? Number(e.target.value) : undefined }))}
                  placeholder="e.g. 80000"
                  className="input"
                />
              </div>
              <div>
                <label htmlFor="salary-max" className="block text-sm text-surface-400 mb-1.5">Maximum</label>
                <input
                  id="salary-max"
                  type="number"
                  value={prefs.salary_max ?? ''}
                  onChange={(e) => setPrefs((p) => ({ ...p, salary_max: e.target.value ? Number(e.target.value) : undefined }))}
                  placeholder="e.g. 150000"
                  className="input"
                />
              </div>
            </div>
          </section>

          {/* Notifications */}
          <section className="glass-card p-6">
            <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
              <Bell className="w-4 h-4 text-brand-400" />
              Notifications
            </h2>
            <div className="space-y-3">
              {[
                { key: 'email_notifications', label: 'Email notifications', desc: 'Get new job alerts by email' },
                { key: 'push_notifications', label: 'Browser push notifications', desc: 'Real-time alerts in your browser' },
              ].map(({ key, label, desc }) => (
                <label key={key} className="flex items-start gap-3 cursor-pointer group">
                  <div
                    onClick={() => setPrefs((p) => ({ ...p, [key]: !p[key as keyof typeof p] }))}
                    className={cn(
                      'w-11 h-6 rounded-full border transition-all flex-shrink-0 relative mt-0.5 cursor-pointer',
                      prefs[key as keyof UserPreferences]
                        ? 'bg-brand-500 border-brand-500'
                        : 'bg-surface-600 border-white/10'
                    )}
                  >
                    <div className={cn(
                      'w-4 h-4 rounded-full bg-white absolute top-0.5 transition-all',
                      prefs[key as keyof UserPreferences] ? 'left-[calc(100%-18px)]' : 'left-0.5'
                    )} />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{label}</p>
                    <p className="text-xs text-surface-400">{desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </section>

          {/* Save button */}
          <div className="flex justify-end">
            <button
              id="save-profile"
              type="submit"
              disabled={saving}
              className={cn('btn-primary px-8 py-3 text-base', saving && 'opacity-60 cursor-not-allowed')}
            >
              {saving ? (
                <span className="flex items-center gap-2">
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  Saving…
                </span>
              ) : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
