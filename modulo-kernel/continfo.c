/* continfo.c - /proc/continfo_so1_<CARNET> (filtro heurístico para "contenedores") */
#include "common.h"

/* Heurística: consideramos "proceso de contenedor" si su cmdline contiene "container" o si el cgroup tiene "docker" o "kubepods".
   Esta heurística puede cambiar dependiendo de tu sistema; ver nota al final sobre detección robusta. */

static bool is_container_task(struct task_struct *task, char *cmdline)
{
    /* 1) cmdline heurístico */
    if (cmdline && (strstr(cmdline, "container") || strstr(cmdline, "docker") || strstr(cmdline, "runc")))
        return true;

    /* 2) revisar cgroup (si está disponible) - sólo heurístico */
#if defined(CONFIG_CGROUPS)
    {
        struct task_cgroup *tc;
        /* task->cgroups no se debe acceder sin protección; usar rcu */
        rcu_read_lock();
        if (task->cgroups) {
            /* intentar comprobar cgroups -> default cgroup name (heurístico) */
            struct cgroup_subsys_state *css;
            /* Tomamos el primer subsystem (cpuset) como heurístico */
            css = task->cgroups->subsys[0];
            if (css && css->cgroup && css->cgroup->kn && css->cgroup->kn->name) {
                const char *name = css->cgroup->kn->name;
                if (strstr(name, "docker") || strstr(name, "kubepods") || strstr(name, "container"))
                {
                    rcu_read_unlock();
                    return true;
                }
            }
        }
        rcu_read_unlock();
    }
#endif
    return false;
}

static int cont_show(struct seq_file *m, void *v)
{
    unsigned long total_kb, free_kb, used_kb;
    struct task_struct *task;

    get_meminfo_kb(&total_kb, &free_kb);
    used_kb = total_kb - free_kb;

    seq_printf(m, "{\n");
    seq_printf(m, "  \"mem_total_kb\": %lu,\n", total_kb);
    seq_printf(m, "  \"mem_free_kb\": %lu,\n", free_kb);
    seq_printf(m, "  \"mem_used_kb\": %lu,\n", used_kb);
    seq_printf(m, "  \"containers\": [\n");

    rcu_read_lock();
    for_each_process(task) {
        unsigned long vsz_kb = 0, rss_kb = 0;
        char cmdline[CMDLINE_MAX] = {0};
        struct mm_struct *mm = task->mm;

        read_task_cmdline(task, cmdline, CMDLINE_MAX);
        if (!is_container_task(task, cmdline)) continue; /* filtra */

        get_mem_from_mm(mm, &vsz_kb, &rss_kb);

        seq_printf(m,
            "    { \"pid\": %d, \"name\": \"%s\", \"cmdline\": \"%s\", \"vsz_kb\": %lu, \"rss_kb\": %lu, \"mem_pct\": %.2f },\n",
            task->pid, task->comm, cmdline, vsz_kb, rss_kb, percent_of(rss_kb, total_kb));
    }
    rcu_read_unlock();

    seq_printf(m, "  ]\n}\n");
    return 0;
}

static int cont_open(struct inode *inode, struct file *file)
{
    return single_open(file, cont_show, NULL);
}

static const struct file_operations cont_fops = {
    .owner = THIS_MODULE,
    .open = cont_open,
    .read = seq_read,
    .llseek = seq_lseek,
    .release = single_release,
};

static int __init cont_init(void)
{
    proc_create(PROC_NAME_CONT, 0, NULL, &cont_fops);
    pr_info("continfo module loaded\n");
    return 0;
}

static void __exit cont_exit(void)
{
    remove_proc_entry(PROC_NAME_CONT, NULL);
    pr_info("continfo module unloaded\n");
}

MODULE_LICENSE("GPL");
MODULE_AUTHOR("JUAN CARLOS CHOC - Proyecto SO1");
module_init(cont_init);
module_exit(cont_exit);
